package tourist

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"pleco-api/internal/ai"
	"pleco-api/internal/cache"
	"pleco-api/internal/contentquality"
	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"
	"pleco-api/internal/search"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AIService       *ai.Service
	AdminAIService  *ai.Service
	DestinationRepo destination.Repository
	EventRepo       event.Repository
	Cache           cache.Store
	SearchClient    *search.Client
}

func NewHandler(aiService *ai.Service, adminAIService *ai.Service, destRepo destination.Repository, eventRepo event.Repository, cacheStore cache.Store, searchClient *search.Client) *Handler {
	if searchClient == nil {
		searchClient = &search.Client{}
	}
	return &Handler{
		AIService:       aiService,
		AdminAIService:  adminAIService,
		DestinationRepo: destRepo,
		EventRepo:       eventRepo,
		Cache:           cacheStore,
		SearchClient:    searchClient,
	}
}

type ChatMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type AIQueryRequest struct {
	Query   string        `json:"query" binding:"required"`
	History []ChatMessage `json:"history"`
}

type AIQueryResponse struct {
	Reply                 string   `json:"reply"`
	MatchedDestinationIDs []string `json:"matchedDestinationIds"`
	MatchedEventIDs       []string `json:"matchedEventIds"`
}

type AIRecommendResponse struct {
	DestinationID string `json:"destinationId"`
	Headline      string `json:"headline"`
	Reason        string `json:"reason"`
	Crowd         string `json:"crowd"`
}

type AIJourneyRequest struct {
	DestinationName string `json:"destinationName" binding:"required"`
}

type AIJourneyResponse struct {
	Steps []JourneyStep `json:"steps"`
}

type JourneyStep struct {
	Time  string `json:"time"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

type AIGenerateDestinationRequest struct {
	DestinationName string `json:"destinationName" binding:"required"`
	Category        string `json:"category"`
	Region          string `json:"region"`
}

type AIFaq struct {
	Q string `json:"q"`
	A string `json:"a"`
}

type AIGenerateDestinationResponse struct {
	Name             string   `json:"name"`
	NameEn           string   `json:"name_en"`
	Category         string   `json:"category"`
	SubRegion        string   `json:"sub_region"`
	Tagline          string   `json:"tagline"`
	TaglineEn        string   `json:"tagline_en"`
	Location         string   `json:"location"`
	Description      string   `json:"description"`
	DescriptionEn    string   `json:"description_en"`
	Story            string   `json:"story"`
	StoryEn          string   `json:"story_en"`
	TicketPrice      string   `json:"ticket_price"`
	OpeningHours     string   `json:"opening_hours"`
	BestTime         string   `json:"best_time"`
	BestTimeEn       string   `json:"best_time_en"`
	Latitude         string   `json:"latitude"`
	Longitude        string   `json:"longitude"`
	Rating           string   `json:"rating"`
	ReviewCount      string   `json:"review_count"`
	Facilities       []string `json:"facilities"`
	FacilitiesEn     []string `json:"facilities_en"`
	TravelTips       []string `json:"travel_tips"`
	TravelTipsEn     []string `json:"travel_tips_en"`
	Faqs             []AIFaq  `json:"faqs"`
	SeoTitle         string   `json:"seo_title"`
	SeoTitleEn       string   `json:"seo_title_en"`
	SeoDescription   string   `json:"seo_description"`
	SeoDescriptionEn string   `json:"seo_description_en"`
	SeoKeywords      string   `json:"seo_keywords"`
	SeoKeywordsEn    string   `json:"seo_keywords_en"`
}

type AIGenerateEventRequest struct {
	EventTitle string `json:"eventTitle"`
	Category   string `json:"category"`
	Location   string `json:"location"`
}

type AIGenerateEventResponse struct {
	Title            string `json:"title"`
	TitleEn          string `json:"title_en"`
	Description      string `json:"description"`
	DescriptionEn    string `json:"description_en"`
	Organizer        string `json:"organizer"`
	TicketPrice      string `json:"ticket_price"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date"`
	MaxAttendees     string `json:"max_attendees"`
	Latitude         string `json:"latitude"`
	Longitude        string `json:"longitude"`
	SeoTitle         string `json:"seo_title"`
	SeoTitleEn       string `json:"seo_title_en"`
	SeoDescription   string `json:"seo_description"`
	SeoDescriptionEn string `json:"seo_description_en"`
	SeoKeywords      string `json:"seo_keywords"`
	SeoKeywordsEn    string `json:"seo_keywords_en"`
}

type AIImageSearchRequest struct {
	Image    string `json:"image" binding:"required"`
	MimeType string `json:"mimeType" binding:"required"`
}

// AITrendingItem represents a single trending pick — can be a destination or event.
type AITrendingItem struct {
	Type      string  `json:"type"`      // "destination" or "event"
	ID        string  `json:"id"`        // external_id of the item
	Badge     string  `json:"badge"`     // e.g. "Spesial Hari Ini", "Trending", "Akan Datang"
	BadgeType string  `json:"badgeType"` // enum: trending, hidden_gem, event, today_only, popular, new, photographers_pick
	Headline  string  `json:"headline"`  // short punchy label
	Reason    string  `json:"reason"`    // one-sentence reason
	ImageURL  string  `json:"imageUrl"`  // thumbnail image
	Rating    float64 `json:"rating"`    // 0 if event
	Distance  string  `json:"distance"`  // e.g. "18 km", empty if unknown
	Location  string  `json:"location"`  // sub-region or event location
}

type AITrendingResponse struct {
	Items []AITrendingItem `json:"items"`
}

func (h *Handler) Trending(c *gin.Context) {
	ctx := context.Background()
	locale := resolveLocale(c)
	isID := locale == "id"

	// ── Cache hit: return stored response without calling AI ─────────────────
	cacheKey := cache.KeyAITrendingResponse(locale)
	if h.Cache != nil {
		var cached AITrendingResponse
		if ok, err := h.Cache.GetJSON(ctx, cacheKey, &cached); err == nil && ok {
			httpx.Success(c, 200, "Trending picks loaded (cached)", cached, nil)
			return
		}
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	events, err := h.EventRepo.FindAll()
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load events")
		return
	}

	// ── AI disabled: return offline fallback ─────────────────────────────────
	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "Trending picks loaded", h.offlineTrendingResponse(dests, events, isID), nil)
		return
	}

	destContext := destinationsContextJSON(dests)
	eventContext := eventsContextJSON(events)

	langInstruction := "Respond in Indonesian (Bahasa Indonesia). All badge, headline, and reason fields must be in Bahasa Indonesia."
	if !isID {
		langInstruction = "Respond in English. All badge, headline, and reason fields must be in English."
	}

	systemInstruction := fmt.Sprintf(`You are an AI tourism curator for Yogyakarta, Indonesia.
Your task is to select exactly 5 trending picks for tourists TODAY. The picks can be a mix of destinations and upcoming events.
Prioritize variety: mix adventure, culture, nature, and events. Make the selection feel fresh and curated.
%s

DESTINATION CATALOG:
%s

UPCOMING EVENTS:
%s

Respond ONLY with valid JSON matching this schema exactly:
{
  "items": [
    {
      "type": "destination" or "event",
      "id": "exact external_id or event id from the catalog",
      "badge": "short badge label in the requested language",
      "badgeType": "one of: trending, hidden_gem, event, today_only, popular, new, photographers_pick",
      "headline": "punchy 3-6 word label",
      "reason": "one sentence why this is trending today",
      "imageUrl": "image URL from the item if available, else empty string",
      "rating": number (destination rating, or 0 for events),
      "distance": "approximate distance string like '18 km' if known, else empty string",
      "location": "sub_region for destinations, location field for events"
    }
  ]
}
badgeType rules:
- "trending" for popular fast-rising picks
- "hidden_gem" for underrated spots
- "event" for upcoming events
- "today_only" for limited-time experiences
- "popular" for all-time favorites
- "new" for newly opened/launched
- "photographers_pick" for most photogenic
Return exactly 5 items. Mix at least 1 event if events are available.`, langInstruction, destContext, eventContext)

	result, err := h.AIService.Generate(ctx, ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   "Pilihkan 5 trending picks terbaik untuk wisatawan di Yogyakarta hari ini.",
		Temperature:  0.65,
		MaxTokens:    1200,
	})
	if err != nil {
		httpx.Success(c, 200, "Trending picks loaded (offline)", h.offlineTrendingResponse(dests, events, isID), nil)
		return
	}

	var parsed AITrendingResponse
	if err := ai.UnmarshalJSON(result.Text, &parsed); err != nil {
		httpx.Success(c, 200, "Trending picks loaded (offline)", h.offlineTrendingResponse(dests, events, isID), nil)
		return
	}

	// Enrich imageUrl from local catalog when AI leaves it empty
	destMap := make(map[string]destination.Destination, len(dests))
	for _, d := range dests {
		destMap[d.ExternalID] = d
	}
	eventMap := make(map[string]event.Event, len(events))
	for _, e := range events {
		eventMap[e.ExternalID] = e
	}

	// Validate and enrich — drop items whose ID doesn't exist in the catalog
	validated := make([]AITrendingItem, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		if item.Type == "destination" {
			if d, ok := destMap[item.ID]; ok {
				item.ImageURL = destImageURL(d)
				if item.Rating == 0 {
					item.Rating = d.Rating
				}
				validated = append(validated, item)
			}
			continue
		}
		if item.Type == "event" {
			if ev, ok := eventMap[item.ID]; ok {
				item.ImageURL = ev.ImageURL
				validated = append(validated, item)
			}
			continue
		}
	}

	// If validation dropped too many, fall back to offline rather than a thin list
	if len(validated) < 3 {
		httpx.Success(c, 200, "Trending picks loaded (offline)", h.offlineTrendingResponse(dests, events, isID), nil)
		return
	}
	parsed.Items = validated

	// ── Save to Redis cache (weekly TTL) ─────────────────────────────────────
	if h.Cache != nil {
		// 1. Full response — used by this endpoint on subsequent calls
		_ = h.Cache.SetJSON(ctx, cacheKey, parsed, cache.TTLAITrending)

		// 2. Just the destination IDs — used by badge logic in destination handler
		var trendingDestIDs []string
		for _, item := range parsed.Items {
			if item.Type == "destination" {
				trendingDestIDs = append(trendingDestIDs, item.ID)
			}
		}
		_ = h.Cache.SetJSON(ctx, cache.KeyAITrendingIDs(locale), trendingDestIDs, cache.TTLAITrending)
	}

	httpx.Success(c, 200, "Trending picks loaded", parsed, nil)
}

// resolveLocale reads Accept-Language header and returns "id" (default) or "en".
func resolveLocale(c *gin.Context) string {
	lang := c.GetHeader("Accept-Language")
	if lang == "" {
		return "id"
	}
	if len(lang) >= 2 && (lang[:2] == "en" || lang[:2] == "EN") {
		return "en"
	}
	return "id"
}

// destImageURL extracts the first image URL from a destination's Images JSON field.
func destImageURL(d destination.Destination) string {
	if len(d.Images) == 0 {
		return ""
	}
	type imgEntry struct {
		URL string `json:"url"`
	}
	var imgs []imgEntry
	if b, err := json.Marshal(d.Images); err == nil {
		_ = json.Unmarshal(b, &imgs)
		if len(imgs) > 0 {
			return imgs[0].URL
		}
	}
	return ""
}

// offlineTrendingResponse builds a curated fallback from the real DB — no AI required.
// It prefers well-known destinations by external_id, then fills remaining slots from
// whatever is in the DB so callers always receive up to 5 items.
func (h *Handler) offlineTrendingResponse(dests []destination.Destination, events []event.Event, isID bool) *AITrendingResponse {
	// Preferred picks with curated badges/headlines — matched by external_id.
	type preferredPick struct {
		id        string
		badge     string
		badgeEN   string
		badgeType string
		head      string
		headEN    string
		why       string
		whyEN     string
	}
	preferred := []preferredPick{
		{"merapi", "Spesial Hari Ini", "Today's Special", "today_only", "Merapi Lava Tour", "Merapi Lava Tour", "Petualangan terbaik di hari yang cerah", "Best adventure on a sunny day"},
		{"prambanan", "Trending", "Trending", "trending", "Prambanan Temple", "Prambanan Temple", "Candi Hindu terbesar di Asia Tenggara", "Largest Hindu temple in Southeast Asia"},
		{"goajomblang", "Hidden Gem", "Hidden Gem", "hidden_gem", "Celestial Beam Cave", "Celestial Beam Cave", "Fenomena cahaya surgawi yang langka", "Rare heavenly light phenomenon"},
		{"tamansari", "Warisan Budaya", "Heritage", "popular", "Taman Sari", "Taman Sari", "Istana air penuh misteri sultan", "Royal water castle full of mystery"},
		{"parangtritis", "Populer", "Popular", "popular", "Pantai Parangtritis", "Parangtritis Beach", "Sunset spektakuler di tepi samudra", "Spectacular ocean sunset"},
	}

	destMap := make(map[string]destination.Destination, len(dests))
	usedIDs := make(map[string]bool)
	for _, d := range dests {
		destMap[d.ExternalID] = d
	}

	items := make([]AITrendingItem, 0, 5)

	// Phase 1: add preferred picks that exist in the DB.
	for _, p := range preferred {
		d, ok := destMap[p.id]
		if !ok {
			continue
		}
		badge, head, why := p.badge, p.head, p.why
		if !isID {
			badge, head, why = p.badgeEN, p.headEN, p.whyEN
		}
		items = append(items, AITrendingItem{
			Type:      "destination",
			ID:        p.id,
			Badge:     badge,
			BadgeType: p.badgeType,
			Headline:  head,
			Reason:    why,
			ImageURL:  destImageURL(d),
			Rating:    d.Rating,
			Location:  d.SubRegion,
		})
		usedIDs[p.id] = true
		if len(items) == 5 {
			break
		}
	}

	// Phase 2: fill remaining slots from DB destinations (highest-rated first).
	if len(items) < 5 {
		badgesID := []string{"Trending", "Populer", "Alam Terbaik", "Warisan Budaya", "Ikon Dunia"}
		badgesEN := []string{"Trending", "Popular", "Best Nature", "Heritage", "World Icon"}
		badgeTypes := []string{"trending", "popular", "hidden_gem", "popular", "trending"}
		badges := badgesEN
		if isID {
			badges = badgesID
		}
		bi := 0
		for _, d := range dests {
			if len(items) == 5 {
				break
			}
			if usedIDs[d.ExternalID] {
				continue
			}
			badge := badges[bi%len(badges)]
			badgeType := badgeTypes[bi%len(badgeTypes)]
			bi++
			items = append(items, AITrendingItem{
				Type:      "destination",
				ID:        d.ExternalID,
				Badge:     badge,
				BadgeType: badgeType,
				Headline:  d.Name,
				Reason:    d.Tagline,
				ImageURL:  destImageURL(d),
				Rating:    d.Rating,
				Location:  d.SubRegion,
			})
			usedIDs[d.ExternalID] = true
		}
	}

	// Phase 3: swap the last destination slot for the nearest upcoming event if available.
	if len(events) > 0 && len(items) > 0 {
		ev := events[0]
		evBadge := "Akan Datang"
		if !isID {
			evBadge = "Upcoming"
		}
		items[len(items)-1] = AITrendingItem{
			Type:      "event",
			ID:        ev.ExternalID,
			Badge:     evBadge,
			BadgeType: "event",
			Headline:  ev.Title,
			Reason:    ev.Description,
			ImageURL:  ev.ImageURL,
			Location:  ev.Location,
		}
	}

	return &AITrendingResponse{Items: items}
}

func (h *Handler) Query(c *gin.Context) {
	var req AIQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline mode", h.offlineQueryResponse(req.Query), nil)
		return
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	eventsData, _ := h.EventRepo.FindAll() // best-effort — don't fail if events unavailable

	destContext := destinationsContextJSON(dests)
	eventContext := eventsContextJSON(eventsData)

	// Pre-query: find events matching the user's query keywords directly from DB.
	// This gives AI a strong hint so it prioritises events over destinations
	// when the query clearly refers to a named event or festival.
	preMatchedEvents, _ := h.EventRepo.Search(req.Query)
	var preMatchedEventIDs []string
	for _, e := range preMatchedEvents {
		preMatchedEventIDs = append(preMatchedEventIDs, e.ExternalID)
	}

	log.Printf("[AI Query] user_query=%q total_destinations=%d total_events=%d pre_matched_events=%v",
		req.Query, len(dests), len(eventsData), preMatchedEventIDs)

	systemInstruction := fmt.Sprintf(`You are a warm, highly knowledgeable, and deeply hospitable local guide from Yogyakarta, Indonesia.
Your task is to act as a "knowledgeable local friend" helping tourists discover destinations and events in Yogyakarta.
Adopt a premium, elegant, yet warm and conversational tone.
Occasionally use gentle Javanese greetings (like 'Sugeng rawuh' for Welcome, 'Matur nuwun' for Thank you, 'Monggo' for Please proceed).

CRITICAL RULES:
1. ALWAYS search the UPCOMING EVENTS catalog first when the user asks about any event, festival, or cultural celebration.
2. If you find ANY matching event — regardless of its status (upcoming, active, completed) or how far in the future it is — you MUST include its "id" field in matchedEventIds.
3. Do NOT filter events by status or date. Show all matching events as-is from the catalog.
4. Do NOT say an event "is not in the list" if it appears in the UPCOMING EVENTS catalog below. Check carefully by title, category, and keywords.
5. matchedEventIds must never be empty if an event matching the query exists in the catalog.
6. For destination questions, include relevant destination IDs in matchedDestinationIds.
7. NEVER mention the catalog, database, data list, or any internal system to the user. Never say phrases like "tidak terdaftar dalam katalog", "tidak ada dalam daftar", "not in our list", "not in the catalog", or any variation. If you don't find something, simply answer warmly based on your knowledge of Yogyakarta without referencing internal data.
8. If PRE-MATCHED EVENT IDs are provided in the user prompt, those events are the highest-priority results — always include ALL of them in matchedEventIds and feature them prominently in the reply.

Here is the exact catalog of Yogyakarta destinations you can recommend. Do NOT invent new places:
%s

UPCOMING EVENTS & FESTIVALS (search this catalog thoroughly for every query):
%s

Respond ONLY with valid JSON matching this schema:
{
  "reply": "Your friendly narrative advice, 3-5 sentences. Include the actual event dates and location from the catalog.",
  "matchedDestinationIds": ["array of destination IDs from the catalog that are relevant, empty array if none"],
  "matchedEventIds": ["array of event IDs from the UPCOMING EVENTS catalog that are relevant, empty array if none"]
}`, destContext, eventContext)

	userPrompt := buildUserPrompt(req.Query, req.History)
	if len(preMatchedEventIDs) > 0 {
		userPrompt += fmt.Sprintf("\n\nPRE-MATCHED EVENT IDs (found by direct keyword search — include ALL in matchedEventIds): %v", preMatchedEventIDs)
	}

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    1500,
	})
	if err != nil {
		httpx.Success(c, 200, "Query processed (offline)", h.offlineQueryResponse(req.Query), nil)
		return
	}

	var parsed AIQueryResponse
	if err := ai.UnmarshalJSON(result.Text, &parsed); err != nil {
		httpx.Success(c, 200, "Query processed (offline)", h.offlineQueryResponse(req.Query), nil)
		return
	}

	log.Printf("[AI Query] result query=%q matched_destinations=%v matched_events=%v",
		req.Query, parsed.MatchedDestinationIDs, parsed.MatchedEventIDs)

	httpx.Success(c, 200, "Query processed", parsed, nil)
}

func (h *Handler) ImageSearch(c *gin.Context) {
	var req AIImageSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline mode", &AIQueryResponse{
			Reply:                 "Sugeng rawuh! In offline mode, I have simulated a local vision scan of your uploaded image. It captures the enchanting heritage and magical energy of Yogyakarta!",
			MatchedDestinationIDs: []string{"tamansari", "prambanan"},
		}, nil)
		return
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	destContext := destinationsContextJSON(dests)
	systemInstruction := fmt.Sprintf(`You are a warm, highly knowledgeable local guide from Yogyakarta, Indonesia.
Analyze the uploaded image and relate it to Yogyakarta's destinations.

Here is our exact catalog of destinations:
%s

Respond ONLY with valid JSON matching this schema:
{
  "reply": "Your narrative describing what you see and matching destinations, 3-5 sentences.",
  "matchedDestinationIds": ["array of matched destination IDs"]
}`, destContext)

	userPrompt := fmt.Sprintf("Image data (base64, mime: %s). Identify if it resembles any tourist attraction in Yogyakarta.", req.MimeType)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    1500,
	})
	if err != nil {
		httpx.Success(c, 200, "Image analyzed (offline)", &AIQueryResponse{
			Reply:                 "Sugeng rawuh! Gambar yang menarik. Berikut beberapa destinasi Yogyakarta yang mungkin relevan.",
			MatchedDestinationIDs: []string{"tamansari", "prambanan"},
		}, nil)
		return
	}

	var parsed AIQueryResponse
	if err := ai.UnmarshalJSON(result.Text, &parsed); err != nil {
		httpx.Success(c, 200, "Image analyzed (offline)", &AIQueryResponse{
			Reply:                 "Sugeng rawuh! Gambar yang menarik. Berikut beberapa destinasi Yogyakarta yang mungkin relevan.",
			MatchedDestinationIDs: []string{"tamansari", "prambanan"},
		}, nil)
		return
	}

	httpx.Success(c, 200, "Image analyzed", parsed, nil)
}

func (h *Handler) Recommend(c *gin.Context) {
	now := fmt.Sprintf("%s", c.Query("time"))
	if now == "" {
		now = "morning"
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline mode", h.offlineRecommendResponse(now), nil)
		return
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	destContext := destinationsContextJSON(dests)
	systemInstruction := fmt.Sprintf(`You are an AI tourism assistant for Yogyakarta, Indonesia.
Pick EXACTLY ONE best destination from the catalog for tourists to visit right now (time of day: %s).
Consider: time of day, typical weather, crowd patterns, and uniqueness of experience.

Here is the exact catalog of Yogyakarta destinations:
%s

Respond ONLY with valid JSON matching this schema:
{
  "destinationId": "the exact id field from the catalog",
  "headline": "A punchy 5-8 word reason why this is the best spot right now (e.g. 'Perfect morning light for temple shots')",
  "reason": "One sentence explaining why this destination is ideal right now.",
  "crowd": "Low" or "Medium" or "High"
}`, now, destContext)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   fmt.Sprintf("Current time of day: %s. Pick the single best destination for a tourist to visit right now.", now),
		Temperature:  0.6,
		MaxTokens:    400,
	})
	if err != nil {
		// AI call failed — return offline fallback instead of error
		httpx.Success(c, 200, "Recommendation generated (offline)", h.offlineRecommendResponse(now), nil)
		return
	}

	var parsed AIRecommendResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil || parsed.DestinationID == "" {
		httpx.Success(c, 200, "Recommendation generated (offline)", h.offlineRecommendResponse(now), nil)
		return
	}

	httpx.Success(c, 200, "Recommendation generated", parsed, nil)
}

func (h *Handler) offlineRecommendResponse(timeOfDay string) *AIRecommendResponse {
	switch {
	case containsAny(timeOfDay, "morning"):
		return &AIRecommendResponse{
			DestinationID: "merapi",
			Headline:      "Perfect morning light for Merapi views",
			Reason:        "Clear morning skies offer the best visibility for Mount Merapi's majestic silhouette.",
			Crowd:         "Low",
		}
	case containsAny(timeOfDay, "afternoon"):
		return &AIRecommendResponse{
			DestinationID: "prambanan",
			Headline:      "Golden afternoon at Prambanan Temple",
			Reason:        "Afternoon light makes the ancient spires glow in warm gold tones.",
			Crowd:         "Medium",
		}
	case containsAny(timeOfDay, "evening", "sunset"):
		return &AIRecommendResponse{
			DestinationID: "parangtritis",
			Headline:      "Magic sunset at Parangtritis Beach",
			Reason:        "The southern coast offers a spectacular sunset over the Indian Ocean every evening.",
			Crowd:         "High",
		}
	default:
		return &AIRecommendResponse{
			DestinationID: "tamansari",
			Headline:      "Explore Taman Sari's hidden tunnels",
			Reason:        "Taman Sari Water Castle is magnificent at any time of day.",
			Crowd:         "Low",
		}
	}
}

func (h *Handler) offlineQueryResponse(query string) *AIQueryResponse {
	lower := query
	reply := "Sugeng rawuh! I am your local Jogja companion. Currently running in offline mode, but I can still guide you! "
	var matchedIDs []string

	switch {
	case containsAny(lower, "sunset", "beach", "ocean", "sea"):
		reply += "I highly recommend visiting Parangtritis Beach for the most magical southern sunset."
		matchedIDs = []string{"parangtritis"}
	case containsAny(lower, "temple", "hindu", "heritage", "history", "ancient", "prambanan"):
		reply += "Prambanan Temple is the absolute pinnacle of Hindu royal architecture."
		matchedIDs = []string{"prambanan"}
	case containsAny(lower, "volcano", "jeep", "offroad", "adventure", "merapi", "mountain", "sunrise"):
		reply += "For a thrilling adventure, head up to Mount Merapi for the Lava Tour!"
		matchedIDs = []string{"merapi"}
	case containsAny(lower, "secret", "hidden", "gem", "cave", "light", "jomblang"):
		reply += "Goa Jomblang is Yogyakarta's ultimate hidden gem."
		matchedIDs = []string{"goajomblang"}
	case containsAny(lower, "bath", "pool", "sultan", "castle", "palace", "taman sari", "tunnel"):
		reply += "Taman Sari Water Castle is a stunning royal retreat."
		matchedIDs = []string{"tamansari"}
	case containsAny(lower, "shop", "batik", "market", "street", "malioboro", "night", "cheap"):
		reply += "Malioboro Street is the living soul of Yogyakarta!"
		matchedIDs = []string{"malioboro"}
	default:
		reply += "Try asking about 'sunset spots', 'adventures', 'temples', or 'hidden caves'!"
		matchedIDs = []string{"prambanan", "malioboro", "parangtritis"}
	}

	return &AIQueryResponse{Reply: reply, MatchedDestinationIDs: matchedIDs}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

func destinationsContextJSON(dests []destination.Destination) string {
	type destSummary struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		Tagline   string  `json:"tagline"`
		Category  string  `json:"category"`
		BestTime  string  `json:"bestTime"`
		SubRegion string  `json:"subRegion"`
		Rating    float64 `json:"rating"`
	}
	// Send all destinations — AI needs full catalog to give accurate recommendations.
	// Summaries are compact so token count stays manageable (~60 tokens per destination).
	summaries := make([]destSummary, len(dests))
	for i, d := range dests {
		summaries[i] = destSummary{
			ID: d.ExternalID, Name: d.Name, Tagline: d.Tagline,
			Category: d.Category, BestTime: d.BestTime, SubRegion: d.SubRegion,
			Rating: d.Rating,
		}
	}
	b, _ := json.Marshal(summaries)
	return string(b)
}

func eventsContextJSON(events []event.Event) string {
	type eventSummary struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Category  string `json:"category"`
		Location  string `json:"location"`
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
		Status    string `json:"status"`
	}
	summaries := make([]eventSummary, len(events))
	for i, e := range events {
		summaries[i] = eventSummary{
			ID: e.ExternalID, Title: e.Title, Category: e.Category,
			Location: e.Location, StartDate: e.StartDate, EndDate: e.EndDate,
			Status: e.Status,
		}
	}
	b, _ := json.Marshal(summaries)
	return string(b)
}

func buildUserPrompt(query string, history []ChatMessage) string {
	if len(history) == 0 {
		return query
	}
	prompt := "Conversation history:\n"
	for _, msg := range history {
		role := "User"
		if msg.Role == "assistant" {
			role = "Guide"
		}
		prompt += fmt.Sprintf("%s: %s\n", role, msg.Text)
	}
	prompt += fmt.Sprintf("User: %s", query)
	return prompt
}

func (h *Handler) Journey(c *gin.Context) {
	var req AIJourneyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline journey fallback", h.offlineJourneyResponse(req.DestinationName), nil)
		return
	}

	systemInstruction := `You are an AI tourism itinerary planner for Yogyakarta, Indonesia.
Your task is to generate a highly cohesive, premium, 3-step daily timeline (itinerary) centered around a single main destination.
Structure the timeline into 3 distinct parts of the day:
1. Morning (around 08:00 - 10:00)
2. Afternoon / Lunch (around 12:00 - 14:00)
3. Late Afternoon / Sunset (around 16:00 - 18:00)

Return ONLY valid JSON matching this schema:
{
  "steps": [
    {
      "time": "HH:MM",
      "title": "A punchy, enticing 4-7 word title",
      "desc": "A descriptive sentence detailing the activities, local vibe, and unique advice (1-2 sentences max)."
    }
  ]
}`

	userPrompt := fmt.Sprintf("Generate a cohesive 3-step daily journey timeline centered around visiting '%s' in Yogyakarta.", req.DestinationName)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    800,
	})
	if err != nil {
		httpx.Success(c, 200, "AI error, using offline journey fallback", h.offlineJourneyResponse(req.DestinationName), nil)
		return
	}

	var parsed AIJourneyResponse
	if err := ai.UnmarshalJSON(result.Text, &parsed); err != nil {
		httpx.Success(c, 200, "AI response parse failure, using offline journey fallback", h.offlineJourneyResponse(req.DestinationName), nil)
		return
	}

	httpx.Success(c, 200, "Journey generated", parsed, nil)
}

func (h *Handler) offlineJourneyResponse(destinationName string) *AIJourneyResponse {
	return &AIJourneyResponse{
		Steps: []JourneyStep{
			{
				Time:  "08:00",
				Title: fmt.Sprintf("Morning Discovery at %s", destinationName),
				Desc:  fmt.Sprintf("Start your journey early to enjoy the cool morning breeze and capture pristine photos of %s.", destinationName),
			},
			{
				Time:  "12:30",
				Title: "Culinary Heritage Lunch",
				Desc:  "Head to a nearby traditional restaurant to savor signature Yogyakarta dishes like Gudeg or hot wedang drinks.",
			},
			{
				Time:  "16:00",
				Title: "Sunset Exploration",
				Desc:  "Wind down your adventure by exploring local handicraft stalls and capturing the beautiful golden hour glow.",
			},
		},
	}
}

func (h *Handler) GenerateDestination(c *gin.Context) {
	var req AIGenerateDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AdminAIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline fallback", h.offlineGenerateDestinationResponse(req.DestinationName, req.Category, req.Region), nil)
		return
	}

	systemInstruction := `You are an expert tourism content writer for Yogyakarta, Indonesia.
Your task is to generate comprehensive, editorial-quality content for the given destination.

GROUNDED RESEARCH:
You may receive a "GROUNDED RESEARCH CONTEXT" block containing verified web-search findings about the destination. You MUST only treat as factual the information present in that context. You MUST NOT fabricate ticket prices, opening hours, coordinates, ratings, or review counts that are not stated there. For any fact not covered by the research context, omit it or leave the field empty rather than inventing it.
If the research context states a ticket price or opening hours, ALWAYS copy that exact information into ticket_price / opening_hours — never leave those fields empty when the fact is present in the context.

CONTENT QUALITY:
- The "story" field should read like a feature article — evocative, narrative-driven, and persuasive
- The "tagline" must be a short, memorable phrase (5-8 words)
- The "description" should be informative yet inviting (2-3 paragraphs)
- Vary your opening sentence between destinations. Avoid generic template phrases.
- Generate SEO-optimized titles, descriptions, and keywords for both Indonesian and English audiences
- Determine the correct category from: Temple, Beach, Nature, Heritage, Cultural, Culinary, Shopping
- Determine the correct administrative region: Sleman, Bantul, Yogyakarta, Gunungkidul, or Kulon Progo (or "Near Yogyakarta" for non-DIY areas like Borobudur)

TONE & VARIETY:
- Warm, conversational Indonesian — like a knowledgeable local friend giving a recommendation, not a government tourism brochure.
- Avoid stiff formal phrases ("menawarkan pengalaman tak terlupakan", "destinasi wisata yang menakjubkan") unless earned by a specific concrete detail.
- Do not open every destination with the same sentence pattern — vary: sometimes a fact, sometimes a sensory detail, sometimes a question.
- For "story_en"/"description_en": same warmth in English, avoid stock phrases like "unforgettable experience" unless followed by a concrete detail.

LENGTH LIMITS (hard caps — never exceed these, they keep the JSON safe):
- description / description_en: 150-250 words each (2-3 short paragraphs)
- story / story_en: 200-320 words each (4-6 sentences)
- tagline / tagline_en: 5-8 words
- facilities: 3-6 short items (1-3 words each)
- travel_tips: 3-5 one-line tips
- faqs: 3-5 FAQ items (question + 1-2 sentence answer, only facts from research or safe general knowledge)
- seo_title: max 60 chars; seo_description: max 160 chars

Return ONLY valid JSON matching this schema:
{
  "name": "Corrected official destination name in Indonesian (use the name the user gave, only fix spelling if needed)",
  "name_en": "English version of the destination name",
  "category": "Factually correct category from: Temple, Beach, Nature, Heritage, Cultural, Culinary, Shopping",
  "sub_region": "One of: Sleman, Bantul, Yogyakarta, Gunungkidul, Kulon Progo, or 'Near Yogyakarta'",
  "tagline": "Short memorable promotional phrase in Indonesian (5-8 words)",
  "tagline_en": "English version of the tagline",
  "location": "Full street address or area description",
  "description": "Informative description in Indonesian (2-3 paragraphs, editorial quality)",
  "description_en": "English version of description",
  "story": "Editorial/feature article in Indonesian about the destination — evocative, narrative-driven, persuasive",
  "story_en": "English version of the editorial story",
  "ticket_price": "Ticket price information with currency, or empty string if unknown",
  "opening_hours": "Opening hours, or empty string if unknown",
  "best_time": "Best time to visit in Indonesian",
  "best_time_en": "Best time to visit in English",
  "latitude": "Decimal latitude coordinate, or empty string if unknown",
  "longitude": "Decimal longitude coordinate, or empty string if unknown",
  "rating": "Average rating as a number string (e.g. '4.3'), or empty string if unknown — NOT invented",
  "review_count": "Total number of reviews as a number string (e.g. '842'), or empty string if unknown — NOT invented",
  "seo_title": "SEO title in Indonesian (max 60 chars)",
  "seo_title_en": "SEO title in English (max 60 chars)",
  "seo_description": "Meta description in Indonesian (max 160 chars)",
  "seo_description_en": "Meta description in English (max 160 chars)",
  "seo_keywords": "Comma-separated keywords in Indonesian",
  "seo_keywords_en": "Comma-separated keywords in English",
  "facilities": ["list of physical amenities in Indonesian (min 3, e.g. Parkir, Toilet, Mushola) — standard amenities of this type of site are fine, only omit if entirely unknown"],
  "facilities_en": ["same amenities in English"],
  "travel_tips": ["practical visitor tips in Indonesian (min 3, e.g. Datang pagi hari) — everyday visitor advice is fine, only omit if entirely unknown"],
  "travel_tips_en": ["same tips in English"],
  "faqs": [{"q": "commonly asked question a visitor would search for, in Indonesian", "a": "concise, accurate answer in Indonesian (1-2 sentences)"}]
}`

	// Ground generation with a web search so the model reasons from real
	// findings instead of fabricating facts/prices/hours/ratings. Fail-open:
	// a search error simply means the prompt is sent ungrounded.
	researchQuery := search.ResearchQuery(req.DestinationName, req.Region)
	researchContext := search.GroundedContext(c.Request.Context(), h.SearchClient, researchQuery)

	baseUserPrompt := fmt.Sprintf(
		"Generate comprehensive destination content for '%s' in Yogyakarta, Indonesia. "+
			"Only use the grounded research context appended above if present; do not invent facts.",
		req.DestinationName,
	)

	var parsed AIGenerateDestinationResponse
	var banned []string
	for attempt := 0; attempt < contentquality.MaxAttempts; attempt++ {
		userPrompt := baseUserPrompt
		if researchContext != "" {
			userPrompt = researchContext + "\n\n" + userPrompt
		}
		if len(banned) > 0 {
			userPrompt = contentquality.Feedback(banned) + "\n\n" + userPrompt
		}

		result, err := h.AdminAIService.Generate(context.Background(), ai.GenerateInput{
			SystemPrompt: systemInstruction,
			UserPrompt:   userPrompt,
			Temperature:  0.7,
			MaxTokens:    8000,
		})
		if err != nil {
			log.Printf("[ai-fallback] GenerateDestination: AI error (%v) for '%s' — offline fallback", err, req.DestinationName)
			httpx.Success(c, 200, "AI error, using offline fallback", h.offlineGenerateDestinationResponse(req.DestinationName, req.Category, req.Region), nil)
			return
		}

		if !parseDestinationPayload(result.Text, &parsed) {
			log.Printf("[ai-fallback] GenerateDestination: parse failure for '%s' — offline fallback. HEAD: %q TAIL: %q", req.DestinationName, result.Text[:min(len(result.Text), 200)], result.Text[max(0, len(result.Text)-200):])
			httpx.Success(c, 200, "AI response parse failure, using offline fallback", h.offlineGenerateDestinationResponse(req.DestinationName, req.Category, req.Region), nil)
			return
		}
		banned = contentquality.FindBanned(contentquality.Prose(
			parsed.Description, parsed.DescriptionEn, parsed.Story, parsed.StoryEn,
			parsed.Tagline, parsed.TaglineEn,
			parsed.SeoTitle, parsed.SeoTitleEn, parsed.SeoDescription, parsed.SeoDescriptionEn,
		))
		if len(banned) == 0 {
			break
		}
		log.Printf("[style-gate] GenerateDestination: clichéd phrase(s) %v for '%s' — regenerating (attempt %d)", banned, req.DestinationName, attempt+2)
	}

	// IMPORTANT: do NOT backfill missing coordinates / ratings from the offline
	// template (root cause #3 — hardcoded -7.7956,110.3695 for every destination,
	// plus fake 4.5/1000). Those facts must come from real research or be filled
	// manually by an admin before publishing. Leaving them empty lets the
	// quality gate (factDensityScore) flag the destination instead of silently
	// publishing fabricated data.

	httpx.Success(c, 200, "Destination content generated", parsed, nil)
}

func (h *Handler) GenerateEvent(c *gin.Context) {
	var req AIGenerateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AdminAIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline fallback", h.offlineGenerateEventResponse(req.EventTitle), nil)
		return
	}

	systemInstruction := fmt.Sprintf(`You are an expert event content writer for Yogyakarta tourism.
Your task is to generate compelling, editorial-quality content for the given event in both Indonesian and English.
You may receive a "GROUNDED RESEARCH CONTEXT" block. You MUST only treat as factual the information in that context. Do NOT invent dates, organizers, ticket prices, or capacity beyond what it states; if a fact is unknown, leave the field empty rather than guessing.
If the research context states a date, organizer, or ticket price, use that exact information — never leave those fields empty when the fact is present in the context.
Today's date is %s.

CRITICAL DATE RULES:
- Use the next upcoming occurrence after today for well-known annual events (Sekaten, Grebeg Maulud, Prambanan Jazz, etc.).
- If you genuinely do not know the real dates for this specific event, return empty strings "" for start_date and end_date. DO NOT guess or invent dates.
- NEVER use today's date as the event date unless the event actually happens today.
- A date like "2026-08-10" for an event clearly scheduled for a different month is WRONG and unacceptable.

CONTENT VARIETY:
- Vary the structure and opening of each event description; avoid repeating the same template phrase across events.

LENGTH LIMITS (hard caps — never exceed these):
- description / description_en: 100-200 words each (1-2 short paragraphs)
- seo_title: max 60 chars; seo_description: max 160 chars

Return ONLY valid JSON matching this schema:
{
  "title": "Corrected event title in Indonesian",
  "title_en": "English version of the event title",
  "description": "Editorial description in Indonesian (compelling, 1-2 paragraphs)",
  "description_en": "English version of the description",
  "organizer": "Organizer name, or empty string if unknown",
  "ticket_price": "Ticket price information, or empty string if unknown",
  "start_date": "Real verified start date in YYYY-MM-DD format, or empty string if unknown",
  "end_date": "Real verified end date in YYYY-MM-DD format, or empty string if unknown",
  "max_attendees": "Estimated max attendees as a plain number string (e.g. '500'), or empty string if unknown",
  "latitude": "Decimal latitude coordinate of the venue, or empty string if unknown",
  "longitude": "Decimal longitude coordinate of the venue, or empty string if unknown",
  "seo_title": "SEO title in Indonesian (max 60 chars)",
  "seo_title_en": "SEO title in English (max 60 chars)",
  "seo_description": "Meta description in Indonesian (max 160 chars)",
  "seo_description_en": "Meta description in English (max 160 chars)",
  "seo_keywords": "Comma-separated keywords in Indonesian",
  "seo_keywords_en": "Comma-separated keywords in English"
}`, time.Now().Format("2006-01-02"))

	// Ground generation with a web search. Fail-open: empty context is fine.
	researchQuery := search.ResearchQuery(req.EventTitle, req.Location)
	researchContext := search.GroundedContext(c.Request.Context(), h.SearchClient, researchQuery)

	baseUserPrompt := fmt.Sprintf(
		"Generate comprehensive event content for '%s' held in %s, category: %s. "+
			"Use only the grounded research context above if present. For dates, use your training knowledge of well-known annual events; if uncertain, leave start_date and end_date as empty strings rather than guessing.",
		req.EventTitle, req.Location, req.Category,
	)

	var parsed AIGenerateEventResponse
	var banned []string
	for attempt := 0; attempt < contentquality.MaxAttempts; attempt++ {
		userPrompt := baseUserPrompt
		if researchContext != "" {
			userPrompt = researchContext + "\n\n" + userPrompt
		}
		if len(banned) > 0 {
			userPrompt = contentquality.Feedback(banned) + "\n\n" + userPrompt
		}

		result, err := h.AdminAIService.Generate(context.Background(), ai.GenerateInput{
			SystemPrompt: systemInstruction,
			UserPrompt:   userPrompt,
			Temperature:  0.7,
			MaxTokens:    4000,
		})
		if err != nil {
			log.Printf("[ai-fallback] GenerateEvent: AI error (%v) for '%s' — offline fallback", err, req.EventTitle)
			httpx.Success(c, 200, "AI error, using offline fallback", h.offlineGenerateEventResponse(req.EventTitle), nil)
			return
		}

		if !parseEventPayload(result.Text, &parsed) {
			log.Printf("[ai-fallback] GenerateEvent: parse failure for '%s' — offline fallback", req.EventTitle)
			httpx.Success(c, 200, "AI response parse failure, using offline fallback", h.offlineGenerateEventResponse(req.EventTitle), nil)
			return
		}

		banned = contentquality.FindBanned(contentquality.Prose(
			parsed.Description, parsed.DescriptionEn,
			parsed.SeoTitle, parsed.SeoTitleEn, parsed.SeoDescription, parsed.SeoDescriptionEn,
		))
		if len(banned) == 0 {
			break
		}
		log.Printf("[style-gate] GenerateEvent: clichéd phrase(s) %v for '%s' — regenerating (attempt %d)", banned, req.EventTitle, attempt+2)
	}

	// Do NOT backfill missing coordinates/max_attendees from the offline template
	// (root cause #3). Leave them empty so the quality gate can flag the item
	// rather than silently publishing fabricated data.

	httpx.Success(c, 200, "Event content generated", parsed, nil)
}

// parseDestinationPayload unmarshals AI destination JSON with a relaxed pass
// that tolerates numeric strings for latitude/longitude/rating/review_count,
// raw control characters inside strings, and truncation that drops the final
// closing brace.
func parseDestinationPayload(text string, parsed *AIGenerateDestinationResponse) bool {
	cleaned := ai.RepairUnterminatedJSON(ai.SanitizeJSONStrings(stripJSONFences(text)))
	if err := ai.UnmarshalJSON(cleaned, parsed); err == nil {
		return true
	}
	var raw map[string]interface{}
	if err := ai.UnmarshalJSON(cleaned, &raw); err != nil {
		return false
	}
	for _, key := range []string{"latitude", "longitude", "rating", "review_count"} {
		if v, ok := raw[key]; ok {
			if n, ok := v.(float64); ok {
				raw[key] = strconv.FormatFloat(n, 'f', -1, 64)
			}
		}
	}
	norm, _ := json.Marshal(raw)
	return json.Unmarshal(norm, parsed) == nil
}

// stripJSONFences removes markdown code fences (```json ... ```) some models
// wrap around JSON despite being asked for raw JSON.
func stripJSONFences(text string) string {
	t := strings.TrimSpace(text)
	if strings.HasPrefix(t, "```") {
		t = strings.TrimPrefix(t, "```")
		if i := strings.IndexByte(t, '\n'); i >= 0 {
			t = t[i+1:]
		}
		t = strings.TrimSpace(t)
		if strings.HasSuffix(t, "```") {
			t = t[:len(t)-3]
		}
	}
	return strings.TrimSpace(t)
}

// parseEventPayload unmarshals AI event JSON with a relaxed pass that tolerates
// numeric strings for max_attendees/latitude/longitude.
func parseEventPayload(text string, parsed *AIGenerateEventResponse) bool {
	if err := ai.UnmarshalJSON(stripJSONFences(text), parsed); err == nil {
		return true
	}
	var raw map[string]interface{}
	if err := ai.UnmarshalJSON(stripJSONFences(text), &raw); err != nil {
		return false
	}
	for _, key := range []string{"max_attendees", "latitude", "longitude"} {
		if v, ok := raw[key]; ok {
			if n, ok := v.(float64); ok {
				raw[key] = strconv.FormatFloat(n, 'f', -1, 64)
			}
		}
	}
	norm, _ := json.Marshal(raw)
	return json.Unmarshal(norm, parsed) == nil
}

func (h *Handler) offlineGenerateEventResponse(title string) *AIGenerateEventResponse {
	return &AIGenerateEventResponse{
		Title:            title,
		TitleEn:          title,
		Description:      fmt.Sprintf("Event '%s' akan diselenggarakan di Yogyakarta. Simak jadwal, lokasi, dan informasi tiket melalui kanal resmi panitia.", title),
		DescriptionEn:    fmt.Sprintf("The event '%s' will be held in Yogyakarta. Check the official organizer channels for schedule, venue, and ticket information.", title),
		Organizer:        "Panitia Event",
		TicketPrice:      "",
		StartDate:        "",
		EndDate:          "",
		MaxAttendees:     "",
		Latitude:         "",
		Longitude:        "",
		SeoTitle:         title + " | JogjaGem",
		SeoTitleEn:       title + " | JogjaGem",
		SeoDescription:   "Informasi terbaru mengenai " + title + ". Dapatkan detail lokasi, jadwal, dan harga tiket di sini.",
		SeoDescriptionEn: "Latest information about " + title + ". Get details on location, schedule, and ticket prices here.",
		SeoKeywords:      title + ", event jogja, wisata jogja",
		SeoKeywordsEn:    title + ", jogja event, jogja tourism",
	}
}

func (h *Handler) offlineGenerateDestinationResponse(name, category, region string) *AIGenerateDestinationResponse {
	cat := category
	if cat == "" {
		cat = "Cultural"
	}
	reg := region
	if reg == "" {
		reg = "Yogyakarta"
	}
	return &AIGenerateDestinationResponse{
		Name:             name,
		NameEn:           name,
		Category:         cat,
		SubRegion:        reg,
		Tagline:          fmt.Sprintf("Temukan Keindahan %s", name),
		TaglineEn:        fmt.Sprintf("Discover the Beauty of %s", name),
		Location:         fmt.Sprintf("%s, Daerah Istimewa Yogyakarta", name),
		Description:      fmt.Sprintf("%s adalah destinasi wisata %s yang menakjubkan di %s, Yogyakarta. Tempat ini menawarkan pengalaman yang tak terlupakan bagi setiap pengunjung dengan keindahan alam dan nilai budayanya yang kaya.", name, cat, reg),
		DescriptionEn:    fmt.Sprintf("%s is a breathtaking %s destination in %s, Yogyakarta. This place offers an unforgettable experience for every visitor with its natural beauty and rich cultural value.", name, cat, reg),
		Story:            fmt.Sprintf("Terletak di %s yang asri, %s menyimpan pesona yang memikat hati setiap pengunjung. Saat pertama kali melangkah, Anda akan disambut oleh suasana yang tenang dan pemandangan yang memanjakan mata. Tempat ini bukan sekadar destinasi wisata, melainkan sebuah perjalanan yang membawa Anda lebih dekat dengan kekayaan budaya dan alam Yogyakarta.", reg, name),
		StoryEn:          fmt.Sprintf("Nestled in the scenic %s, %s captivates the heart of every visitor. As you step in, you are greeted by a tranquil atmosphere and breathtaking views. This is not just a tourism destination, but a journey that brings you closer to Yogyakarta's rich culture and nature.", reg, name),
		TicketPrice:      "",
		OpeningHours:     "",
		BestTime:         "Pagi hari atau sore menjelang matahari terbenam",
		BestTimeEn:       "Early morning or late afternoon near sunset",
		Latitude:         "",
		Longitude:        "",
		SeoTitle:         fmt.Sprintf("%s - Wisata %s di %s | Jogjagem", name, cat, reg),
		SeoTitleEn:       fmt.Sprintf("%s - %s Tourism in %s | Jogjagem", name, cat, reg),
		SeoDescription:   fmt.Sprintf("Kunjungi %s, destinasi %s terbaik di %s. Nikmati pengalaman wisata yang tak terlupakan.", name, cat, reg),
		SeoDescriptionEn: fmt.Sprintf("Visit %s, the best %s destination in %s. Enjoy an unforgettable tourism experience.", name, cat, reg),
		Rating:           "",
		ReviewCount:      "",
		SeoKeywords:      fmt.Sprintf("%s, %s, %s, Wisata Jogja, Destinasi Wisata", name, cat, reg),
		SeoKeywordsEn:    fmt.Sprintf("%s, %s, %s, Jogja Tourism, Tourist Destination", name, cat, reg),
	}
}

// AIMultiRecommendItem is a single pick in the multi-recommendation response.
type AIMultiRecommendItem struct {
	DestinationID string  `json:"destinationId"`
	Headline      string  `json:"headline"`
	Reason        string  `json:"reason"`
	Badge         string  `json:"badge"`
	Crowd         string  `json:"crowd"`
	ImageURL      string  `json:"imageUrl"`
	Rating        float64 `json:"rating"`
	Location      string  `json:"location"`
}

type AIMultiRecommendResponse struct {
	Items []AIMultiRecommendItem `json:"items"`
}

// MultiRecommend returns 2-4 AI-curated destination picks with variety (different categories).
func (h *Handler) MultiRecommend(c *gin.Context) {
	now := c.Query("time")
	if now == "" {
		now = "morning"
	}

	locale := resolveLocale(c)
	isID := locale == "id"

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	destMap := make(map[string]destination.Destination, len(dests))
	for _, d := range dests {
		destMap[d.ExternalID] = d
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline picks", h.offlineMultiRecommend(now, dests, isID), nil)
		return
	}

	langInstruction := "Respond in Indonesian (Bahasa Indonesia)."
	if !isID {
		langInstruction = "Respond in English."
	}

	destContext := destinationsContextJSON(dests)
	systemInstruction := fmt.Sprintf(`You are an AI tourism curator for Yogyakarta, Indonesia.
Select EXACTLY 4 destinations from the catalog to display in an "AI Picks Just for You" section.
Time of day: %s.
%s

Rules:
- Pick destinations from DIFFERENT categories (e.g. nature, heritage, beach, adventure, hidden-gem, culinary)
- Vary the crowd levels (Low / Medium / High)
- Assign a short punchy badge per pick (e.g. "AI Pick Today", "Hidden Gem", "Sunset Spot", "Adventure Call")
- Make picks feel fresh and curated for right now
- headline and reason MUST be in the requested language

DESTINATION CATALOG:
%s

Respond ONLY with valid JSON:
{
  "items": [
    {
      "destinationId": "exact id from catalog",
      "headline": "punchy 4-6 word label",
      "reason": "one sentence why now",
      "badge": "short badge label",
      "crowd": "Low" or "Medium" or "High",
      "imageUrl": "image URL from catalog or empty string",
      "rating": number,
      "location": "sub_region from catalog"
    }
  ]
}
Return exactly 4 items.`, now, langInstruction, destContext)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   fmt.Sprintf("Time of day: %s. Pick 4 diverse AI-curated destinations for tourists right now.", now),
		Temperature:  0.7,
		MaxTokens:    900,
	})
	if err != nil {
		httpx.Success(c, 200, "Picks loaded (offline)", h.offlineMultiRecommend(now, dests, isID), nil)
		return
	}

	var parsed AIMultiRecommendResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil || len(parsed.Items) == 0 {
		httpx.Success(c, 200, "Picks loaded (offline)", h.offlineMultiRecommend(now, dests, isID), nil)
		return
	}

	// Validate and enrich — drop items whose ID doesn't exist in the catalog
	validated := make([]AIMultiRecommendItem, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		if d, ok := destMap[item.DestinationID]; ok {
			item.ImageURL = destImageURL(d)
			if item.Rating == 0 {
				item.Rating = d.Rating
			}
			if item.Location == "" {
				item.Location = d.SubRegion
			}
			validated = append(validated, item)
		}
	}

	if len(validated) < 2 {
		httpx.Success(c, 200, "Picks loaded (offline)", h.offlineMultiRecommend(now, dests, isID), nil)
		return
	}
	parsed.Items = validated

	httpx.Success(c, 200, "AI picks loaded", parsed, nil)
}

// offlineMultiRecommend returns curated fallback picks without calling the AI.
func (h *Handler) offlineMultiRecommend(timeOfDay string, dests []destination.Destination, isID bool) *AIMultiRecommendResponse {
	type pick struct {
		id      string
		badge   string
		badgeEN string
		head    string
		headEN  string
		why     string
		whyEN   string
	}

	var ordered []pick
	switch {
	case containsAny(timeOfDay, "morning"):
		ordered = []pick{
			{"merapi", "Pilihan AI Hari Ini", "AI Pick Today", "Merapi Sunrise Jeep Tour", "Merapi Sunrise Jeep Tour", "Pemandangan terbaik gunung berapi aktif di pagi hari", "Best morning views of the active volcano"},
			{"goajomblang", "Hidden Gem", "Hidden Gem", "Goa Jomblang", "Goa Jomblang", "Kolom cahaya surga yang langka di pagi hari", "Rare heavenly light column at peak morning"},
			{"prambanan", "Warisan Budaya", "Heritage Gem", "Candi Prambanan", "Prambanan Temple", "Cahaya emas pagi di puncak candi kuno", "Golden morning light on ancient spires"},
			{"kalibiru", "Alam Terbaik", "Nature Pick", "Hutan Kalibiru", "Kalibiru Forest", "Jalan setapak berkabut di antara kanopi pohon", "Misty canopy walks at their best"},
		}
	case containsAny(timeOfDay, "afternoon"):
		ordered = []pick{
			{"prambanan", "Pilihan AI Hari Ini", "AI Pick Today", "Candi Prambanan", "Prambanan Temple", "Cahaya sore yang hangat di puncak candi Hindu", "Warm afternoon glow on Hindu spires"},
			{"tamansari", "Warisan Budaya", "Heritage Pick", "Taman Sari", "Taman Sari Castle", "Jelajahi terowongan kerajaan di siang hari", "Afternoon exploration of royal tunnels"},
			{"ratuboko", "Siap Sunset", "Sunset Prep", "Istana Ratu Boko", "Ratu Boko Palace", "Spot terbaik menunggu golden hour tiba", "Prime spot to wait for the golden hour"},
			{"pindul", "Petualangan", "Adventure Call", "Goa Pindul", "Goa Pindul", "Cave tubing yang menyegarkan di sore hari", "Refreshing cave tubing in afternoon coolness"},
		}
	default:
		ordered = []pick{
			{"parangtritis", "Spot Sunset", "Sunset Spot", "Pantai Parangtritis", "Parangtritis Beach", "Sunset spektakuler di atas Samudra Hindia", "Spectacular Indian Ocean sunset"},
			{"tamansari", "Pilihan AI Hari Ini", "AI Pick Today", "Taman Sari", "Taman Sari Castle", "Suasana mistis di malam hari", "Mystical evening atmosphere"},
			{"malioboro", "Malam Meriah", "Night Vibes", "Jalan Malioboro", "Malioboro Street", "Kuliner jalanan dan budaya di malam hari", "Vibrant evening street food and culture"},
			{"tebingbreksi", "Hidden Gem", "Hidden Gem", "Tebing Breksi", "Tebing Breksi", "Tebing dramatis diterangi sinar matahari terbenam", "Dramatic cliffs lit by the setting sun"},
		}
	}

	destMap := make(map[string]destination.Destination, len(dests))
	for _, d := range dests {
		destMap[d.ExternalID] = d
	}

	items := make([]AIMultiRecommendItem, 0, 4)
	for _, p := range ordered {
		d, ok := destMap[p.id]
		if !ok {
			continue
		}
		badge := p.badgeEN
		head := p.headEN
		why := p.whyEN
		if isID {
			badge = p.badge
			head = p.head
			why = p.why
		}
		items = append(items, AIMultiRecommendItem{
			DestinationID: p.id,
			Headline:      head,
			Reason:        why,
			Badge:         badge,
			Crowd:         "Low",
			ImageURL:      destImageURL(d),
			Rating:        d.Rating,
			Location:      d.SubRegion,
		})
	}

	// Fill remaining from DB if preferred IDs not found
	if len(items) < 4 {
		used := make(map[string]bool)
		for _, item := range items {
			used[item.DestinationID] = true
		}
		badgesID := []string{"Trending", "Populer", "Alam Terbaik", "Warisan Budaya"}
		badgesEN := []string{"Trending", "Popular", "Best Nature", "Cultural Heritage"}
		bi := 0
		for _, d := range dests {
			if len(items) >= 4 {
				break
			}
			if used[d.ExternalID] {
				continue
			}
			badge := badgesEN[bi%len(badgesEN)]
			reason := d.Tagline
			if isID {
				badge = badgesID[bi%len(badgesID)]
			}
			items = append(items, AIMultiRecommendItem{
				DestinationID: d.ExternalID,
				Headline:      d.Name,
				Reason:        reason,
				Badge:         badge,
				Crowd:         "Low",
				ImageURL:      destImageURL(d),
				Rating:        d.Rating,
				Location:      d.SubRegion,
			})
			used[d.ExternalID] = true
			bi++
		}
	}

	return &AIMultiRecommendResponse{Items: items}
}

// RouteTimelineNode is a single slot in the rolling 4-slot route.
type RouteTimelineNode struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
	Location    string  `json:"location"`
	SubRegion   string  `json:"subRegion"`
	Rating      float64 `json:"rating"`
	DistanceKm  float64 `json:"distanceKm"`
	IsPast      bool    `json:"isPast"`
	IsCurrent   bool    `json:"isCurrent"`
	IsTomorrow  bool    `json:"isTomorrow"`
	DayLabel    string  `json:"dayLabel"`
	DisplayTime string  `json:"displayTime"`
	TimeSlot    string  `json:"timeSlot"`
	Duration    string  `json:"duration"`
}

type RouteTimelineResponse struct {
	HeaderTitle string              `json:"headerTitle"`
	TimeRange   string              `json:"timeRange"`
	Nodes       []RouteTimelineNode `json:"nodes"`
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	if lat1 == 0 || lon1 == 0 || lat2 == 0 || lon2 == 0 {
		return 1.5
	}
	const R = 6371.0
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2.0)*math.Sin(dLat/2.0) +
		math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*
			math.Sin(dLon/2.0)*math.Sin(dLon/2.0)
	return R * 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
}

func isFarRegionJump(reg1, reg2 string) bool {
	r1 := strings.ToLower(reg1)
	r2 := strings.ToLower(reg2)
	if (strings.Contains(r1, "gunungkidul") && strings.Contains(r2, "kulon")) ||
		(strings.Contains(r1, "kulon") && strings.Contains(r2, "gunungkidul")) {
		return true
	}
	return false
}

// RouteTimeline returns 4 sequential rolling timeline slots starting from current time.
func (h *Handler) RouteTimeline(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	hourStr := c.Query("hour")
	savedIDsStr := c.Query("saved_ids")
	categoryFilter := strings.ToLower(strings.TrimSpace(c.Query("category")))

	savedMap := make(map[string]bool)
	if savedIDsStr != "" {
		for _, id := range strings.Split(savedIDsStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				savedMap[id] = true
			}
		}
	}

	userLat := -7.7828
	userLng := 110.3671
	if v, err := strconv.ParseFloat(latStr, 64); err == nil && v != 0 {
		userLat = v
	}
	if v, err := strconv.ParseFloat(lngStr, 64); err == nil && v != 0 {
		userLng = v
	}

	currentHour := time.Now().Hour()
	if hVal, err := strconv.Atoi(hourStr); err == nil && hVal >= 0 && hVal <= 23 {
		currentHour = hVal
	}

	startPeriod := 0
	if currentHour >= 5 && currentHour < 11 {
		startPeriod = 0 // Pagi (07.00 AM)
	} else if currentHour >= 11 && currentHour < 15 {
		startPeriod = 1 // Siang (12.00 PM)
	} else if currentHour == 15 {
		startPeriod = 2 // Sore (03.30 PM)
	} else {
		startPeriod = 3 // Malam (07.30 PM) - 16:00 (4 PM+) onwards
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil || len(dests) == 0 {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	slots := []struct {
		Name       string
		Time       string
		Duration   string
		Categories []string
	}{
		{Name: "Pagi", Time: "07.00 AM", Duration: "~2.5 jam", Categories: []string{"heritage", "nature", "adventure"}},
		{Name: "Siang", Time: "12.00 PM", Duration: "~1.5 jam", Categories: []string{"culinary", "cultural"}},
		{Name: "Sore", Time: "03.30 PM", Duration: "~2.5 jam", Categories: []string{"nature", "beach", "sunset", "hidden-gem"}},
		{Name: "Malam", Time: "07.30 PM", Duration: "~3 jam", Categories: []string{"night_vibes", "culinary", "heritage", "cultural", "shopping"}},
	}

	usedIDs := make(map[string]bool)
	var nodes []RouteTimelineNode

	refLat := userLat
	refLng := userLng
	refSubRegion := ""

	for offset := 0; offset < 4; offset++ {
		periodIdx := (startPeriod + offset) % 4
		isTomorrow := (startPeriod + offset) >= 4
		slot := slots[periodIdx]

		type scoredDest struct {
			dest  destination.Destination
			score float64
			dist  float64
		}
		var candidates []scoredDest

		for _, d := range dests {
			if usedIDs[d.ExternalID] {
				continue
			}
			distFromRef := haversineKm(refLat, refLng, d.Latitude, d.Longitude)

			// Distance penalty: 1 km = -1.2 score points
			score := (d.Rating * 2.5) - (distFromRef * 1.2)

			// Heavy penalty for jumping across opposite ends of Yogyakarta (Gunungkidul <-> Kulon Progo)
			if refSubRegion != "" && isFarRegionJump(refSubRegion, d.SubRegion) {
				score -= 300.0
			}

			if isTomorrow {
				// Priority 1: Wishlist / Saved items get top priority
				if savedMap[d.ExternalID] {
					score += 500.0
				}
				// Priority 2: Mood category filter match bonus (balanced so it respects geographic proximity!)
				if categoryFilter != "" && categoryFilter != "all" {
					catLower := strings.ToLower(d.Category)
					if strings.Contains(catLower, categoryFilter) || strings.Contains(categoryFilter, catLower) {
						score += 20.0
					}
				}
			}

			candidates = append(candidates, scoredDest{dest: d, score: score, dist: distFromRef})
		}

		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].score > candidates[j].score
		})

		var chosen destination.Destination
		if len(candidates) > 0 {
			chosen = candidates[0].dest
		} else {
			chosen = dests[0]
		}
		usedIDs[chosen.ExternalID] = true

		// Calculate step-by-step leg distance (from previous node coordinates to chosen node)
		stepDist := haversineKm(refLat, refLng, chosen.Latitude, chosen.Longitude)

		// Update reference location & subregion for next chained node
		if chosen.Latitude != 0 && chosen.Longitude != 0 {
			refLat = chosen.Latitude
			refLng = chosen.Longitude
		}
		if chosen.SubRegion != "" {
			refSubRegion = chosen.SubRegion
		}

		dayLabel := "HARI INI"
		displayTime := slot.Time
		if isTomorrow {
			dayLabel = "BESOK"
			displayTime = "Besok " + slot.Time
		}

		nodes = append(nodes, RouteTimelineNode{
			ID:          chosen.ExternalID,
			Title:       chosen.Name,
			Type:        "destination",
			Category:    chosen.Category,
			Image:       destImageURL(chosen),
			Location:    chosen.SubRegion,
			SubRegion:   chosen.SubRegion,
			Rating:      chosen.Rating,
			DistanceKm:  stepDist,
			IsPast:      false,
			IsCurrent:   offset == 0,
			IsTomorrow:  isTomorrow,
			DayLabel:    dayLabel,
			DisplayTime: displayTime,
			TimeSlot:    slot.Name,
			Duration:    slot.Duration,
		})
	}

	timeRange := nodes[0].DisplayTime + " → " + nodes[3].DisplayTime

	httpx.Success(c, 200, "Route timeline retrieved successfully", RouteTimelineResponse{
		HeaderTitle: "Rute & Timeline",
		TimeRange:   timeRange,
		Nodes:       nodes,
	}, nil)
}

// NextStopNode is the response for a single next-destination request.
type NextStopNode struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Category     string  `json:"category"`
	Image        string  `json:"image"`
	Location     string  `json:"location"`
	SubRegion    string  `json:"subRegion"`
	Rating       float64 `json:"rating"`
	DistanceKm   float64 `json:"distanceKm"`
	TimeWarning  string  `json:"timeWarning,omitempty"`
	IsTomorrow   bool    `json:"isTomorrow"`
	ScheduledFor string  `json:"scheduledFor,omitempty"`
}

func matchMoodCategory(dest destination.Destination, filter string) bool {
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" || filter == "all" || filter == "semua" {
		return true
	}

	cat := strings.ToLower(strings.TrimSpace(dest.Category))
	name := strings.ToLower(strings.TrimSpace(dest.Name))
	combined := name + " " + cat

	if filter == "nature" || filter == "alam" {
		// Explicit exclusions: if it's craft, weaving, cultural, temple, or museum -> NOT nature!
		if strings.Contains(combined, "budaya") || strings.Contains(combined, "tenun") ||
			strings.Contains(combined, "batik") || strings.Contains(combined, "candi") ||
			strings.Contains(combined, "situs") || strings.Contains(combined, "museum") ||
			strings.Contains(combined, "kerajinan") {
			return false
		}
		return strings.Contains(cat, "nature") || strings.Contains(cat, "alam") ||
			strings.Contains(combined, "bukit") || strings.Contains(combined, "gunung") ||
			strings.Contains(combined, "air terjun") || strings.Contains(combined, "hutan") ||
			strings.Contains(combined, "curug") || strings.Contains(combined, "embung") ||
			strings.Contains(combined, "gua") || strings.Contains(combined, "puncak") ||
			strings.Contains(combined, "tebing") || strings.Contains(combined, "pantai") ||
			strings.Contains(combined, "beach")
	}

	if filter == "pantai" || filter == "beach" {
		return strings.Contains(combined, "pantai") || strings.Contains(combined, "beach") || strings.Contains(combined, "bahari")
	}

	if filter == "cultural" || filter == "budaya" || filter == "heritage" {
		return strings.Contains(combined, "cultural") || strings.Contains(combined, "budaya") ||
			strings.Contains(combined, "heritage") || strings.Contains(combined, "candi") ||
			strings.Contains(combined, "sejarah") || strings.Contains(combined, "museum") ||
			strings.Contains(combined, "situs") || strings.Contains(combined, "tenun") ||
			strings.Contains(combined, "batik") || strings.Contains(combined, "keraton") ||
			strings.Contains(combined, "tamansari") || strings.Contains(combined, "kerajinan")
	}

	if filter == "culinary" || filter == "kuliner" || filter == "food" {
		return strings.Contains(combined, "culinary") || strings.Contains(combined, "kuliner") ||
			strings.Contains(combined, "food") || strings.Contains(combined, "makanan") ||
			strings.Contains(combined, "kopi") || strings.Contains(combined, "gudeg") ||
			strings.Contains(combined, "sate") || strings.Contains(combined, "warung") ||
			strings.Contains(combined, "resto") || strings.Contains(combined, "soto")
	}

	return strings.Contains(combined, filter) || strings.Contains(filter, combined)
}

// NextStop resolves a single next destination near the given ref coordinates, filtered by mood category and hour of day.
// GET /ai/next-stop?lat=...&lng=...&category=...&exclude=id1,id2,...&hour=18
func (h *Handler) NextStop(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")
	categoryFilter := strings.ToLower(strings.TrimSpace(c.Query("category")))
	excludeStr := c.Query("exclude")
	hourStr := c.Query("hour")

	hour := time.Now().Hour()
	if hVal, err := strconv.Atoi(hourStr); err == nil && hVal >= 0 && hVal <= 23 {
		hour = hVal
	}

	refLat := -7.7828
	refLng := 110.3671
	if v, err := strconv.ParseFloat(latStr, 64); err == nil && v != 0 {
		refLat = v
	}
	if v, err := strconv.ParseFloat(lngStr, 64); err == nil && v != 0 {
		refLng = v
	}

	excludeMap := make(map[string]bool)
	if excludeStr != "" {
		for _, id := range strings.Split(excludeStr, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				excludeMap[id] = true
			}
		}
	}

	dests, err := h.DestinationRepo.FindAll("")
	if err != nil || len(dests) == 0 {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	// First filter by requested mood category strictly if passed
	var matchingDests []destination.Destination
	if categoryFilter != "" && categoryFilter != "all" && categoryFilter != "semua" {
		for _, d := range dests {
			if excludeMap[d.ExternalID] {
				continue
			}
			if matchMoodCategory(d, categoryFilter) {
				matchingDests = append(matchingDests, d)
			}
		}
	}

	// Fallback to all unexcluded dests if no specific match
	if len(matchingDests) == 0 {
		for _, d := range dests {
			if !excludeMap[d.ExternalID] {
				matchingDests = append(matchingDests, d)
			}
		}
	}

	type scored struct {
		dest  destination.Destination
		score float64
	}
	var candidates []scored

	for _, d := range matchingDests {
		dist := haversineKm(refLat, refLng, d.Latitude, d.Longitude)
		score := (d.Rating * 2.5) - (dist * 1.2)

		cat := strings.ToLower(d.Category)

		// Apply time-based penalties for closed categories after 17:00 (5 PM)
		if hour >= 17 && (strings.Contains(cat, "nature") || strings.Contains(cat, "alam") || strings.Contains(cat, "pantai") || strings.Contains(cat, "beach")) {
			score -= 150.0
		}
		if hour >= 17 && (strings.Contains(cat, "cultural") || strings.Contains(cat, "budaya") || strings.Contains(cat, "heritage") || strings.Contains(cat, "candi") || strings.Contains(cat, "sejarah") || strings.Contains(cat, "museum")) {
			score -= 150.0
		}

		candidates = append(candidates, scored{dest: d, score: score})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) == 0 {
		httpx.ErrorWithCode(c, 404, "NOT_FOUND", "No destination found")
		return
	}

	chosen := candidates[0].dest
	dist := haversineKm(refLat, refLng, chosen.Latitude, chosen.Longitude)

	isTomorrow := false
	scheduledFor := ""
	timeWarning := ""

	// Use the chosen destination's actual category, not the request filter,
	// because fallback may select a different category than what was requested.
	chosenCat := strings.ToLower(chosen.Category)
	isNatureOrBeach := strings.Contains(chosenCat, "nature") || strings.Contains(chosenCat, "alam") || strings.Contains(chosenCat, "pantai") || strings.Contains(chosenCat, "beach") || strings.Contains(chosenCat, "bukit")
	isCultural := strings.Contains(chosenCat, "cultural") || strings.Contains(chosenCat, "budaya") || strings.Contains(chosenCat, "heritage") || strings.Contains(chosenCat, "candi") || strings.Contains(chosenCat, "sejarah") || strings.Contains(chosenCat, "museum")

	if hour >= 17 && isNatureOrBeach {
		isTomorrow = true
		scheduledFor = "Besok Pagi (07:00)"
		timeWarning = "Destinasi pantai/alam umumnya tutup atau kurang aman setelah jam 17:00. Kami jadwalkan untuk besok pagi."
	} else if hour >= 17 && isCultural {
		isTomorrow = true
		scheduledFor = "Besok Siang (12:00)"
		timeWarning = "Situs budaya/candi/sejarah umumnya tutup setelah jam 17:00. Kami jadwalkan untuk besok."
	}

	httpx.Success(c, 200, "Next stop resolved", NextStopNode{
		ID:           chosen.ExternalID,
		Title:        chosen.Name,
		Category:     chosen.Category,
		Image:        destImageURL(chosen),
		Location:     chosen.SubRegion,
		SubRegion:    chosen.SubRegion,
		Rating:       chosen.Rating,
		DistanceKm:   dist,
		TimeWarning:  timeWarning,
		IsTomorrow:   isTomorrow,
		ScheduledFor: scheduledFor,
	}, nil)
}

func (h *Handler) GenerateArticle(c *gin.Context) {
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.ValidationError(c, httpx.FormatValidationError(err))
		return
	}

	if !h.AdminAIService.Enabled() {
		httpx.ErrorWithCode(c, 503, "AI_DISABLED", "AI service is currently disabled")
		return
	}

	systemPrompt := `You are an expert travel writer for Yogyakarta, Indonesia. Write a comprehensive, engaging, and informative travel article based on the provided title.
	The article should be in BOTH Indonesian and English, editorial quality.
	
	Return ONLY valid JSON in this schema:
	{
	  "title": "...",
	  "titleEn": "...",
	  "content": "HTML formatted article body in Indonesian...",
	  "contentEn": "HTML formatted article body in English...",
	  "excerpt": "...",
	  "excerptEn": "...",
	  "seoTitle": "...",
	  "seoTitleEn": "...",
	  "seoDescription": "...",
	  "seoDescriptionEn": "...",
	  "seoKeywords": "...",
	  "seoKeywordsEn": "..."
	}`

	userPrompt := fmt.Sprintf("Write an article with the title: %s", req.Title)

	result, err := h.AdminAIService.Generate(c.Request.Context(), ai.GenerateInput{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
	})

	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	// Parse the JSON string from AI and return it directly
	var aiResponse map[string]interface{}
	if err := ai.UnmarshalJSON(result.Text, &aiResponse); err != nil {
		httpx.ErrorWithCode(c, 500, "AI_PARSING_ERROR", "Failed to parse AI response")
		return
	}

	httpx.Success(c, 200, "Article generated", aiResponse, nil)
}
