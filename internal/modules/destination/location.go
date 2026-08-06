package destination

import (
	"strings"

	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
)

// ValidRegions is the canonical list of DIY kabupaten/kota + the Near Yogyakarta bucket.
// Matches the same list used in business registration (business/service.go).
var ValidRegions = []string{
	"kota-yogyakarta",
	"sleman",
	"bantul",
	"kulon-progo",
	"gunungkidul",
	"near-yogyakarta",
}

// regionSlugToName converts a URL slug to the canonical SubRegion value stored in the DB.
var regionSlugToName = map[string]string{
	"kota-yogyakarta": "Yogyakarta",
	"sleman":          "Sleman",
	"bantul":          "Bantul",
	"kulon-progo":     "Kulon Progo",
	"gunungkidul":     "Gunungkidul",
	"near-yogyakarta": "Near Yogyakarta",
}

// regionNameToSlug is the reverse map for sitemap generation.
var regionNameToSlug = map[string]string{
	"Yogyakarta":      "kota-yogyakarta",
	"Kota Yogyakarta": "kota-yogyakarta",
	"Sleman":          "sleman",
	"Bantul":          "bantul",
	"Kulon Progo":     "kulon-progo",
	"Gunungkidul":     "gunungkidul",
	"Near Yogyakarta": "near-yogyakarta",
}

// RegionSlugToName is exported for use in the sitemap and frontend API routes.
func RegionSlugToName(slug string) (string, bool) {
	v, ok := regionSlugToName[strings.ToLower(slug)]
	return v, ok
}

// RegionNameToSlug converts a DB SubRegion value to its URL slug equivalent.
func RegionNameToSlug(name string) string {
	if slug, ok := regionNameToSlug[name]; ok {
		return slug
	}
	// Fallback: lowercase + hyphenate
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

// LocationSummary is the response shape for GET /locations/:region
type LocationSummary struct {
	Region       string                `json:"region"`
	RegionSlug   string                `json:"region_slug"`
	TotalCount   int                   `json:"total_count"`
	Destinations []DestinationResponse `json:"destinations"`
	Categories   []CategoryCount       `json:"categories"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// LocationHandler handles /locations/* routes.
type LocationHandler struct {
	DestHandler *Handler
	DestService *Service
	DestRepo    Repository
}

func NewLocationHandler(destHandler *Handler, destService *Service, destRepo Repository) *LocationHandler {
	return &LocationHandler{DestHandler: destHandler, DestService: destService, DestRepo: destRepo}
}

// GET /locations — list all regions with destination counts
func (h *LocationHandler) ListRegions(c *gin.Context) {
	dests, err := h.DestService.GetAll("published")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	type regionSummary struct {
		Slug  string `json:"slug"`
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	counts := make(map[string]int)
	for _, d := range dests {
		counts[d.SubRegion]++
	}

	regions := make([]regionSummary, 0, len(regionNameToSlug))
	for name, slug := range regionNameToSlug {
		if c, ok := counts[name]; ok && c > 0 {
			regions = append(regions, regionSummary{Slug: slug, Name: name, Count: c})
		}
	}

	httpx.Success(c, 200, "Regions fetched", regions, nil)
}

// GET /locations/:region — destinations grouped by a single SubRegion
func (h *LocationHandler) GetByRegion(c *gin.Context) {
	locale := resolveLocale(c)
	regionSlug := strings.ToLower(c.Param("region"))

	regionName, ok := RegionSlugToName(regionSlug)
	if !ok {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "Region not found")
		return
	}

	dests, err := h.DestRepo.FindByRegion(regionName, "published")
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	trendingIDs := h.DestHandler.loadTrendingIDs(locale)
	catCounts := make(map[string]int)
	responses := make([]DestinationResponse, 0, len(dests))

	for _, d := range dests {
		localized := d.Localize(locale)
		responses = append(responses, localized.ToResponse(trendingIDs))
		catCounts[d.Category]++
	}

	categories := make([]CategoryCount, 0, len(catCounts))
	for cat, count := range catCounts {
		categories = append(categories, CategoryCount{Category: cat, Count: count})
	}

	summary := LocationSummary{
		Region:       regionName,
		RegionSlug:   regionSlug,
		TotalCount:   len(responses),
		Destinations: responses,
		Categories:   categories,
	}
	httpx.Success(c, 200, "Location destinations fetched", summary, nil)
}

// SetupLocationRoutes registers /locations endpoints.
func SetupLocationRoutes(api *gin.RouterGroup, handler *LocationHandler) {
	loc := api.Group("/locations")
	loc.GET("", handler.ListRegions)
	loc.GET("/:region", handler.GetByRegion)
}
