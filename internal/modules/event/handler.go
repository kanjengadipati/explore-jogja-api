package event

import (
	"sort"

	"pleco-api/internal/cache"
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Service *Service
	Cache   cache.Store
}

func NewHandler(service *Service, cacheStore cache.Store) *Handler {
	return &Handler{
		Service: service,
		Cache:   cacheStore,
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	cacheKey := cache.KeyEventsAll

	pag := httpx.ParsePagination(c)
	if c.Query("limit") == "" {
		pag.Limit = 15 // override default to 15
	}

	var allResponses []EventResponse
	cacheHit := false
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &allResponses); err == nil && ok {
		cacheHit = true
	}

	if !cacheHit {
		trendingIDs := loadTrendingIDs(h.Cache)
		events, err := h.Service.GetAll()
		if err != nil {
			httpx.HandleError(c, err)
			return
		}

		allResponses = make([]EventResponse, len(events))
		for i, e := range events {
			allResponses[i] = e.ToResponse(trendingIDs)
		}

		// Sort by badge priority: trending (0), populer (1), terbatas (2), akan_datang (3), others (4)
		// Within the same badge, sort by StartDate DESC (newest events first), then ID ASC.
		badgeRank := func(badge string) int {
			switch badge {
			case "trending":
				return 0
			case "populer":
				return 1
			case "terbatas":
				return 2
			case "akan_datang":
				return 3
			default:
				return 4
			}
		}
		sort.SliceStable(allResponses, func(i, j int) bool {
			ri, rj := badgeRank(allResponses[i].Badge), badgeRank(allResponses[j].Badge)
			if ri != rj {
				return ri < rj
			}
			if allResponses[i].StartDate != allResponses[j].StartDate {
				return allResponses[i].StartDate > allResponses[j].StartDate
			}
			return allResponses[i].ID < allResponses[j].ID
		})

		_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, allResponses, cache.TTLEvents)
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
	httpx.Success(c, 200, "Events fetched", paged, meta)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	cacheKey := cache.KeyEventsIDPrefix + id

	var cachedResponse EventResponse
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &cachedResponse); err == nil && ok {
		httpx.Success(c, 200, "Event fetched (cached)", cachedResponse, nil)
		return
	}

	trendingIDs := loadTrendingIDs(h.Cache)
	event, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Event not found")
		return
	}

	response := event.ToResponse(trendingIDs)

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, response, cache.TTLEvents)
	httpx.Success(c, 200, "Event fetched", response, nil)
}

func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}

	trendingIDs := loadTrendingIDs(h.Cache)
	events, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	responses := make([]EventResponse, len(events))
	for i, e := range events {
		responses[i] = e.ToResponse(trendingIDs)
	}

	httpx.Success(c, 200, "Search results", responses, nil)
}

func (h *Handler) Create(c *gin.Context) {
	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}
	if err := h.Service.Create(&event); err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Invalidate cache
	ctx := c.Request.Context()
	_ = h.Cache.Delete(ctx, cache.KeyEventsAll)

	httpx.Success(c, 201, "Event created", event, nil)
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Invalid request body")
		return
	}

	event, err := h.Service.Update(id, req)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Invalidate cache
	ctx := c.Request.Context()
	_ = h.Cache.Delete(ctx, cache.KeyEventsAll, cache.KeyEventsIDPrefix+id)

	httpx.Success(c, 200, "Event updated", event, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Invalidate cache
	ctx := c.Request.Context()
	_ = h.Cache.Delete(ctx, cache.KeyEventsAll, cache.KeyEventsIDPrefix+id)

	httpx.Success(c, 200, "Event deleted", nil, nil)
}
