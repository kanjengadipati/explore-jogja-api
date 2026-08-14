package event

import (
	"sort"
	"strings"
	"time"

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

// resolveLocale reads Accept-Language header and returns "id" (default) or "en".
func resolveLocale(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "id"
	}
	if strings.HasPrefix(strings.ToLower(lang), "en") {
		return "en"
	}
	return "id"
}

func (h *Handler) GetAll(c *gin.Context) {
	locale := resolveLocale(c)
	cacheKey := cache.KeyEventsAll(locale)

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
		trendingIDs := loadTrendingIDs(h.Cache, locale)
		events, err := h.Service.GetAll()
		if err != nil {
			httpx.HandleError(c, err)
			return
		}

		allResponses = make([]EventResponse, len(events))
		for i, e := range events {
			allResponses[i] = e.ToResponse(locale, trendingIDs)
		}

		// Sort by event lifecycle relative to today: active (0) → upcoming (1) →
		// completed (2) → unknown/no-date (3). Within each group events nearest
		// to today get priority; badge priority is the final tiebreaker.
		today := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02")
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
			a, b := allResponses[i], allResponses[j]
			ra, rb := eventStatusRank(a.StartDate, a.EndDate, today), eventStatusRank(b.StartDate, b.EndDate, today)
			if ra != rb {
				return ra < rb
			}
			// Near-date priority within each lifecycle group.
			switch ra {
			case 0: // active: least time left first
				if a.EndDate != b.EndDate {
					return a.EndDate < b.EndDate
				}
				if a.StartDate != b.StartDate {
					return a.StartDate > b.StartDate
				}
			case 1: // upcoming: nearest start first
				if a.StartDate != b.StartDate {
					return a.StartDate < b.StartDate
				}
				if a.EndDate != b.EndDate {
					return a.EndDate < b.EndDate
				}
			case 2: // completed: most recently ended first
				if a.EndDate != b.EndDate {
					return a.EndDate > b.EndDate
				}
				if a.StartDate != b.StartDate {
					return a.StartDate > b.StartDate
				}
			}
			ri, rj := badgeRank(a.Badge), badgeRank(b.Badge)
			if ri != rj {
				return ri < rj
			}
			return a.ID < b.ID
		})

		_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, allResponses, cache.TTLEvents)
	}

	// Server-side filtering on the full sorted list so filter counts and
	// pagination reflect the entire dataset, not just the loaded page.
	category := strings.ToLower(strings.TrimSpace(c.Query("category")))
	q := strings.ToLower(strings.TrimSpace(c.Query("q")))

	if category != "" {
		filtered := allResponses[:0]
		for _, e := range allResponses {
			if strings.ToLower(e.Category) == category {
				filtered = append(filtered, e)
			}
		}
		allResponses = filtered
	}

	if q != "" {
		filtered := allResponses[:0]
		for _, e := range allResponses {
			if strings.Contains(strings.ToLower(e.Title), q) ||
				strings.Contains(strings.ToLower(e.Description), q) ||
				strings.Contains(strings.ToLower(e.Location), q) {
				filtered = append(filtered, e)
			}
		}
		allResponses = filtered
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
	locale := resolveLocale(c)
	cacheKey := cache.KeyEventsID(locale, id)

	var cachedResponse EventResponse
	if ok, err := h.Cache.GetJSON(c.Request.Context(), cacheKey, &cachedResponse); err == nil && ok {
		httpx.Success(c, 200, "Event fetched (cached)", cachedResponse, nil)
		return
	}

	trendingIDs := loadTrendingIDs(h.Cache, locale)
	event, err := h.Service.GetByID(id)
	if err != nil {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Event not found")
		return
	}

	response := event.ToResponse(locale, trendingIDs)

	_ = h.Cache.SetJSON(c.Request.Context(), cacheKey, response, cache.TTLEvents)
	httpx.Success(c, 200, "Event fetched", response, nil)
}

func (h *Handler) Search(c *gin.Context) {
	locale := resolveLocale(c)
	query := c.Query("q")
	if query == "" {
		httpx.ErrorWithCode(c, 400, "VALIDATION_FAILED", "Query parameter 'q' is required")
		return
	}

	trendingIDs := loadTrendingIDs(h.Cache, locale)
	events, err := h.Service.Search(query)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	responses := make([]EventResponse, len(events))
	for i, e := range events {
		responses[i] = e.ToResponse(locale, trendingIDs)
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

	// Invalidate cache (all locale variants)
	ctx := c.Request.Context()
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsAllPrefix)
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsIDAllPrefix)

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

	// Invalidate cache (all locale variants)
	ctx := c.Request.Context()
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsAllPrefix)
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsIDAllPrefix)

	httpx.Success(c, 200, "Event updated", event, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.Service.Delete(id); err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Invalidate cache (all locale variants)
	ctx := c.Request.Context()
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsAllPrefix)
	_ = h.Cache.DeletePrefix(ctx, cache.KeyEventsIDAllPrefix)

	httpx.Success(c, 200, "Event deleted", nil, nil)
}

// eventStatusRank classifies an event relative to today using its date window:
//
//	0 = active   (start <= today <= end)
//	1 = upcoming (start > today)
//	2 = completed (end < today)
//	3 = unknown   (no usable start date)
//
// Dates are YYYY-MM-DD strings so plain string comparison works.
func eventStatusRank(start, end, today string) int {
	if strings.TrimSpace(start) == "" {
		return 3
	}
	if strings.TrimSpace(end) == "" {
		if start <= today {
			return 0 // started and no end date → assume still ongoing
		}
		return 1 // not started yet → upcoming
	}
	if start <= today && end >= today {
		return 0
	}
	if start > today {
		return 1
	}
	return 2
}
