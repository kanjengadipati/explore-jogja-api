package tourist

import (
	"context"
	"encoding/json"
	"fmt"

	"pleco-api/internal/ai"
	"pleco-api/internal/cache"
	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AIService       *ai.Service
	DestinationRepo destination.Repository
	EventRepo       event.Repository
	Cache           cache.Store
}

func NewHandler(aiService *ai.Service, destRepo destination.Repository, eventRepo event.Repository, cacheStore cache.Store) *Handler {
	return &Handler{
		AIService:       aiService,
		DestinationRepo: destRepo,
		EventRepo:       eventRepo,
		Cache:           cacheStore,
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
	Reply               string   `json:"reply"`
	MatchedDestinationIDs []string `json:"matchedDestinationIds"`
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

type AIImageSearchRequest struct {
	Image    string `json:"image" binding:"required"`
	MimeType string `json:"mimeType" binding:"required"`
}

// AITrendingItem represents a single trending pick — can be a destination or event.
type AITrendingItem struct {
	Type      string `json:"type"`      // "destination" or "event"
	ID        string `json:"id"`        // external_id of the item
	Badge     string `json:"badge"`     // e.g. "Spesial Hari Ini", "Trending", "Akan Datang"
	Headline  string `json:"headline"`  // short punchy label
	Reason    string `json:"reason"`    // one-sentence reason
	ImageURL  string `json:"imageUrl"`  // thumbnail image
	Rating    float64 `json:"rating"`   // 0 if event
	Distance  string `json:"distance"`  // e.g. "18 km", empty if unknown
	Location  string `json:"location"`  // sub-region or event location
}

type AITrendingResponse struct {
	Items []AITrendingItem `json:"items"`
}

func (h *Handler) Trending(c *gin.Context) {
	ctx := context.Background()

	// ── Cache hit: return stored response without calling AI ─────────────────
	if h.Cache != nil {
		var cached AITrendingResponse
		if ok, err := h.Cache.GetJSON(ctx, cache.KeyAITrendingResponse, &cached); err == nil && ok {
			httpx.Success(c, 200, "Trending picks loaded (cached)", cached, nil)
			return
		}
	}

	dests, err := h.DestinationRepo.FindAll()
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
		httpx.Success(c, 200, "Trending picks loaded", h.offlineTrendingResponse(dests, events), nil)
		return
	}

	destContext := destinationsContextJSON(dests)
	eventContext := eventsContextJSON(events)

	systemInstruction := fmt.Sprintf(`You are an AI tourism curator for Yogyakarta, Indonesia.
Your task is to select exactly 5 trending picks for tourists TODAY. The picks can be a mix of destinations and upcoming events.
Prioritize variety: mix adventure, culture, nature, and events. Make the selection feel fresh and curated.

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
      "badge": "short badge label in Indonesian, e.g. Spesial Hari Ini / Trending / Hidden Gem / Akan Datang / Populer",
      "headline": "punchy 3-6 word Indonesian or English label",
      "reason": "one sentence why this is trending today",
      "imageUrl": "image URL from the item if available, else empty string",
      "rating": number (destination rating, or 0 for events),
      "distance": "approximate distance string like '18 km' if known, else empty string",
      "location": "sub_region for destinations, location field for events"
    }
  ]
}
Return exactly 5 items. Mix at least 1 event if events are available.`, destContext, eventContext)

	result, err := h.AIService.Generate(ctx, ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   "Pilihkan 5 trending picks terbaik untuk wisatawan di Yogyakarta hari ini.",
		Temperature:  0.65,
		MaxTokens:    1200,
	})
	if err != nil {
		httpx.Success(c, 200, "Trending picks loaded (offline)", h.offlineTrendingResponse(dests, events), nil)
		return
	}

	var parsed AITrendingResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		httpx.Success(c, 200, "Trending picks loaded (offline)", h.offlineTrendingResponse(dests, events), nil)
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

	for i, item := range parsed.Items {
		// Always replace imageUrl from local catalog — AI often hallucinates fake URLs
		if item.Type == "destination" {
			if d, ok := destMap[item.ID]; ok {
				parsed.Items[i].ImageURL = destImageURL(d)
			}
		} else if item.Type == "event" {
			if ev, ok := eventMap[item.ID]; ok {
				parsed.Items[i].ImageURL = ev.ImageURL
			}
		}
		// Enrich rating for destinations
		if item.Type == "destination" && item.Rating == 0 {
			if d, ok := destMap[item.ID]; ok {
				parsed.Items[i].Rating = d.Rating
			}
		}
	}

	// ── Save to Redis cache (weekly TTL) ─────────────────────────────────────
	if h.Cache != nil {
		// 1. Full response — used by this endpoint on subsequent calls
		_ = h.Cache.SetJSON(ctx, cache.KeyAITrendingResponse, parsed, cache.TTLAITrending)

		// 2. Just the destination IDs — used by badge logic in destination handler
		var trendingDestIDs []string
		for _, item := range parsed.Items {
			if item.Type == "destination" {
				trendingDestIDs = append(trendingDestIDs, item.ID)
			}
		}
		_ = h.Cache.SetJSON(ctx, cache.KeyAITrendingIDs, trendingDestIDs, cache.TTLAITrending)
	}

	httpx.Success(c, 200, "Trending picks loaded", parsed, nil)
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
func (h *Handler) offlineTrendingResponse(dests []destination.Destination, events []event.Event) *AITrendingResponse {
	// Preferred picks with curated badges/headlines — matched by external_id.
	preferred := []struct {
		id    string
		badge string
		head  string
		why   string
	}{
		{"merapi", "Spesial Hari Ini", "Merapi Lava Tour", "Petualangan terbaik di hari yang cerah"},
		{"prambanan", "Trending", "Prambanan Temple", "Candi Hindu terbesar di Asia Tenggara"},
		{"goajomblang", "Hidden Gem", "Celestial Beam Cave", "Fenomena cahaya surgawi yang langka"},
		{"tamansari", "Warisan Budaya", "Taman Sari", "Istana air penuh misteri sultan"},
		{"parangtritis", "Populer", "Pantai Parangtritis", "Sunset spektakuler di tepi samudra"},
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
		items = append(items, AITrendingItem{
			Type:     "destination",
			ID:       p.id,
			Badge:    p.badge,
			Headline: p.head,
			Reason:   p.why,
			ImageURL: destImageURL(d),
			Rating:   d.Rating,
			Location: d.SubRegion,
		})
		usedIDs[p.id] = true
		if len(items) == 5 {
			break
		}
	}

	// Phase 2: fill remaining slots from DB destinations (highest-rated first).
	if len(items) < 5 {
		badges := []string{"Trending", "Populer", "Alam Terbaik", "Warisan Budaya", "Ikon Dunia"}
		bi := 0
		for _, d := range dests {
			if len(items) == 5 {
				break
			}
			if usedIDs[d.ExternalID] {
				continue
			}
			badge := badges[bi%len(badges)]
			bi++
			items = append(items, AITrendingItem{
				Type:     "destination",
				ID:       d.ExternalID,
				Badge:    badge,
				Headline: d.Name,
				Reason:   d.Tagline,
				ImageURL: destImageURL(d),
				Rating:   d.Rating,
				Location: d.SubRegion,
			})
			usedIDs[d.ExternalID] = true
		}
	}

	// Phase 3: swap the last destination slot for the nearest upcoming event if available.
	if len(events) > 0 && len(items) > 0 {
		ev := events[0]
		items[len(items)-1] = AITrendingItem{
			Type:     "event",
			ID:       ev.ExternalID,
			Badge:    "Akan Datang",
			Headline: ev.Title,
			Reason:   ev.Description,
			ImageURL: ev.ImageURL,
			Location: ev.Location,
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

	dests, err := h.DestinationRepo.FindAll()
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	eventsData, _ := h.EventRepo.FindAll() // best-effort — don't fail if events unavailable

	destContext  := destinationsContextJSON(dests)
	eventContext := eventsContextJSON(eventsData)

	systemInstruction := fmt.Sprintf(`You are a warm, highly knowledgeable, and deeply hospitable local guide from Yogyakarta, Indonesia.
Your task is to act as a "knowledgeable local friend" helping tourists discover destinations and events in Yogyakarta.
Adopt a premium, elegant, yet warm and conversational tone.
Occasionally use gentle Javanese greetings (like 'Sugeng rawuh' for Welcome, 'Matur nuwun' for Thank you, 'Monggo' for Please proceed).
Answer inquiries thoroughly and recommend specific places from the list of actual destinations provided.
If the user asks about events or festivals, refer to the UPCOMING EVENTS catalog.

Here is the exact catalog of Yogyakarta destinations you can recommend. Do NOT invent new places; map the user's request intelligently to these options:
%s

UPCOMING EVENTS & FESTIVALS:
%s

Respond ONLY with valid JSON matching this schema:
{
  "reply": "Your friendly narrative advice, 3-5 sentences.",
  "matchedDestinationIds": ["array of destination IDs from the catalog that are relevant"]
}`, destContext, eventContext)

	userPrompt := buildUserPrompt(req.Query, req.History)

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
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		httpx.Success(c, 200, "Query processed (offline)", h.offlineQueryResponse(req.Query), nil)
		return
	}

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
			Reply:               "Sugeng rawuh! In offline mode, I have simulated a local vision scan of your uploaded image. It captures the enchanting heritage and magical energy of Yogyakarta!",
			MatchedDestinationIDs: []string{"tamansari", "prambanan"},
		}, nil)
		return
	}

	dests, err := h.DestinationRepo.FindAll()
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
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
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

	dests, err := h.DestinationRepo.FindAll()
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
	limit := len(events)
	if limit > 15 {
		limit = 15
	}
	summaries := make([]eventSummary, limit)
	for i := 0; i < limit; i++ {
		e := events[i]
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
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
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

// AIMultiRecommendItem is a single pick in the multi-recommendation response.
type AIMultiRecommendItem struct {
	DestinationID string `json:"destinationId"`
	Headline      string `json:"headline"`
	Reason        string `json:"reason"`
	Badge         string `json:"badge"`
	Crowd         string `json:"crowd"`
	ImageURL      string `json:"imageUrl"`
	Rating        float64 `json:"rating"`
	Location      string `json:"location"`
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

	dests, err := h.DestinationRepo.FindAll()
	if err != nil {
		httpx.ErrorWithCode(c, 500, "SERVER_INTERNAL_ERROR", "Failed to load destinations")
		return
	}

	destMap := make(map[string]destination.Destination, len(dests))
	for _, d := range dests {
		destMap[d.ExternalID] = d
	}

	if !h.AIService.Enabled() {
		httpx.Success(c, 200, "AI disabled, using offline picks", h.offlineMultiRecommend(now, dests), nil)
		return
	}

	destContext := destinationsContextJSON(dests)
	systemInstruction := fmt.Sprintf(`You are an AI tourism curator for Yogyakarta, Indonesia.
Select EXACTLY 4 destinations from the catalog to display in an "AI Picks Just for You" section.
Time of day: %s.

Rules:
- Pick destinations from DIFFERENT categories (e.g. nature, heritage, beach, adventure, hidden-gem, culinary)
- Vary the crowd levels (Low / Medium / High)
- Assign a short punchy badge per pick (e.g. "AI Pick Today", "Hidden Gem", "Sunset Spot", "Adventure Call")
- Make picks feel fresh and curated for right now

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
Return exactly 4 items.`, now, destContext)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   fmt.Sprintf("Time of day: %s. Pick 4 diverse AI-curated destinations for tourists right now.", now),
		Temperature:  0.7,
		MaxTokens:    900,
	})
	if err != nil {
		httpx.Success(c, 200, "Picks loaded (offline)", h.offlineMultiRecommend(now, dests), nil)
		return
	}

	var parsed AIMultiRecommendResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil || len(parsed.Items) == 0 {
		httpx.Success(c, 200, "Picks loaded (offline)", h.offlineMultiRecommend(now, dests), nil)
		return
	}

	// Always replace imageUrl from local catalog — AI often hallucinates fake URLs
	for i, item := range parsed.Items {
		if d, ok := destMap[item.DestinationID]; ok {
			parsed.Items[i].ImageURL = destImageURL(d)
			if item.Rating == 0 {
				parsed.Items[i].Rating = d.Rating
			}
			if item.Location == "" {
				parsed.Items[i].Location = d.SubRegion
			}
		}
	}

	httpx.Success(c, 200, "AI picks loaded", parsed, nil)
}

// offlineMultiRecommend returns curated fallback picks without calling the AI.
func (h *Handler) offlineMultiRecommend(timeOfDay string, dests []destination.Destination) *AIMultiRecommendResponse {
	type pick struct {
		id    string
		badge string
		head  string
		why   string
	}

	var ordered []pick
	switch {
	case containsAny(timeOfDay, "morning"):
		ordered = []pick{
			{"merapi", "AI Pick Today", "Merapi Sunrise Jeep Tour", "Best morning views of the active volcano"},
			{"goajomblang", "Hidden Gem", "Celestial Beam Cave", "Rare heavenly light column at peak morning"},
			{"prambanan", "Heritage Gem", "Prambanan Temple", "Golden morning light on ancient spires"},
			{"kalibiru", "Nature Pick", "Kalibiru Forest", "Misty canopy walks at their best"},
		}
	case containsAny(timeOfDay, "afternoon"):
		ordered = []pick{
			{"prambanan", "AI Pick Today", "Prambanan Temple", "Warm afternoon glow on Hindu spires"},
			{"tamansari", "Heritage Pick", "Taman Sari Castle", "Afternoon exploration of royal tunnels"},
			{"ratuboko", "Sunset Prep", "Ratu Boko Palace", "Prime spot to wait for the golden hour"},
			{"pindul", "Adventure Call", "Goa Pindul", "Refreshing cave tubing in afternoon coolness"},
		}
	default: // evening / sunset / night
		ordered = []pick{
			{"parangtritis", "Sunset Spot", "Parangtritis Beach", "Spectacular Indian Ocean sunset"},
			{"tamansari", "AI Pick Today", "Taman Sari Castle", "Mystical evening atmosphere"},
			{"malioboro", "Night Vibes", "Malioboro Street", "Vibrant evening street food and culture"},
			{"tebingbreksi", "Hidden Gem", "Tebing Breksi", "Dramatic cliffs lit by the setting sun"},
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
		items = append(items, AIMultiRecommendItem{
			DestinationID: p.id,
			Headline:      p.head,
			Reason:        p.why,
			Badge:         p.badge,
			Crowd:         "Low",
			ImageURL:      destImageURL(d),
			Rating:        d.Rating,
			Location:      d.SubRegion,
		})
	}

	// If preferred IDs weren't in DB, fill from first available destinations
	if len(items) < 4 {
		used := make(map[string]bool)
		for _, item := range items {
			used[item.DestinationID] = true
		}
		badges := []string{"Trending", "Populer", "Alam Terbaik", "Warisan Budaya"}
		bi := 0
		for _, d := range dests {
			if len(items) >= 4 {
				break
			}
			if used[d.ExternalID] {
				continue
			}
			items = append(items, AIMultiRecommendItem{
				DestinationID: d.ExternalID,
				Headline:      d.Name,
				Reason:        d.Tagline,
				Badge:         badges[bi%len(badges)],
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
