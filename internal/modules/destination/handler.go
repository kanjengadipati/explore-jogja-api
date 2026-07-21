package destination

import (
	"context"
	"sort"
	"strings"

	"pleco-api/internal/cache"
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
	Cache   cache.Store
}

func NewHandler(service *Service, cacheStore cache.Store) *Handler {
	return &Handler{Service: service, Cache: cacheStore}
}

// resolveLocale reads the Accept-Language header and returns "id" (default) or "en".
func resolveLocale(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "id"
	}
	lang = strings.ToLower(strings.TrimSpace(strings.Split(lang, ",")[0]))
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "id"
}

// loadTrendingIDs reads the AI-selected trending destination IDs from Redis.
// Returns an empty map (never nil) when the key is missing or Redis is unavailable.
func (h *Handler) loadTrendingIDs() map[string]bool {
	var ids []string
	ok, err := h.Cache.GetJSON(context.Background(), cache.KeyAITrendingIDs, &ids)
	if err != nil || !ok || len(ids) == 0 {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func (h *Handler) GetAll(c *gin.Context) {
	locale := resolveLocale(c)
	cacheKey := cache.KeyDestinationsAllPrefix + locale

	// Parse pagination — default limit 15, max 100
	pag := httpx.ParsePagination(c)
	if c.Query("limit") == "" {
		pag.Limit = 15 // override default to 15
	}

	// Try to serve from cache (full sorted list stored once)
	var allResponses []DestinationResponse
	cacheHit := false
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &allResponses); err == nil && ok {
		cacheHit = true
	}

	if !cacheHit {
		trendingIDs := h.loadTrendingIDs()
		dests, err := h.Service.GetAll()
		if err != nil {
			httpx.HandleError(c, err)
			return
		}
		allResponses = make([]DestinationResponse, len(dests))
		for i, d := range dests {
			localized := d.Localize(locale)
			allResponses[i] = localized.ToResponse(trendingIDs)
		}

		// Sort: trending first, then hidden_gem, then by rating DESC
		badgeRank := func(badge string) int {
			switch badge {
			case "trending":
				return 0
			case "hidden_gem":
				return 1
			default:
				return 2
			}
		}
		sort.SliceStable(allResponses, func(i, j int) bool {
			ri, rj := badgeRank(string(allResponses[i].Badge)), badgeRank(string(allResponses[j].Badge))
			if ri != rj {
				return ri < rj
			}
			return allResponses[i].Rating > allResponses[j].Rating
		})

		_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, allResponses, cache.TTLDestinations)
	}

	// Apply pagination on the full sorted list
	total := int64(len(allResponses))
	start := pag.Offset
	if start > len(allResponses) {
		start = len(allResponses)
	}
	end := start + pag.Limit
	if end > len(allResponses) {
		end = len(allResponses)
	}
	paged := allResponses[start:end]

	meta := httpx.BuildPaginationMeta(total, pag.Page(), pag.Limit)
	httpx.Success(c, 200, "Destinations fetched", paged, meta)
}

func (h *Handler) GetByID(c *gin.Context) {
	locale := resolveLocale(c)
	id := c.Param("id")
	cacheKey := cache.KeyDestinationsIDPrefix + id + ":" + locale

	var cachedResponse DestinationResponse
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &cachedResponse); err == nil && ok {
		httpx.Success(c, 200, "Destination fetched (cached)", cachedResponse, nil)
		return
	}

	trendingIDs := h.loadTrendingIDs()
	dest, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Destination not found")
		return
	}
	localized := dest.Localize(locale)
	response := localized.ToResponse(trendingIDs)

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, response, cache.TTLDestinations)
	httpx.Success(c, 200, "Destination fetched", response, nil)
}

func (h *Handler) GetByCategory(c *gin.Context) {
	locale := resolveLocale(c)
	category := c.Param("category")
	cacheKey := cache.KeyDestinationsCategoryPrefix + category + ":" + locale

	var cachedResponses []DestinationResponse
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &cachedResponses); err == nil && ok {
		httpx.Success(c, 200, "Destinations fetched (cached)", cachedResponses, nil)
		return
	}

	trendingIDs := h.loadTrendingIDs()
	dests, err := h.Service.GetByCategory(category)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	responses := make([]DestinationResponse, len(dests))
	for i, d := range dests {
		localized := d.Localize(locale)
		responses[i] = localized.ToResponse(trendingIDs)
	}

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, responses, cache.TTLDestinations)
	httpx.Success(c, 200, "Destinations fetched", responses, nil)
}


func (h *Handler) Search(c *gin.Context) {
	locale := resolveLocale(c)
	trendingIDs := h.loadTrendingIDs()
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}
	dests, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	responses := make([]DestinationResponse, len(dests))
	for i, d := range dests {
		localized := d.Localize(locale)
		responses[i] = localized.ToResponse(trendingIDs)
	}
	httpx.Success(c, 200, "Search results", responses, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	dest, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Invalidate cache
	ctx := c.Request.Context()
	_ = h.Cache.DeletePrefix(ctx, cache.KeyDestinationsAllPrefix)
	_ = h.Cache.DeletePrefix(ctx, cache.KeyDestinationsCategoryPrefix)
	_ = h.Cache.Delete(ctx, cache.KeyDestinationsIDPrefix+id+":id", cache.KeyDestinationsIDPrefix+id+":en")

	httpx.Success(c, 200, "Destination updated", dest, nil)
}

func (h *Handler) UpdateUserDestinationStatus(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}
	slug := c.Param("slug")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Error(c, 400, "Invalid request")
		return
	}
	if err := h.Service.UpdateUserDestination(userID, slug, req.Status); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Status updated", nil, nil)
}

func (h *Handler) GetUserDestinations(c *gin.Context) {
	userID, ok := httpx.GetUserIDFromContext(c)
	if !ok {
		httpx.Error(c, 401, "Unauthorized")
		return
	}
	uds, err := h.Service.GetUserDestinations(userID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "User destinations fetched", uds, nil)
}
