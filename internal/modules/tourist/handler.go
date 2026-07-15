package tourist

import (
	"context"
	"encoding/json"
	"fmt"

	"pleco-api/internal/ai"
	"pleco-api/internal/httpx"
	"pleco-api/internal/modules/destination"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	AIService         *ai.Service
	DestinationRepo   destination.Repository
}

func NewHandler(aiService *ai.Service, destRepo destination.Repository) *Handler {
	return &Handler{
		AIService:       aiService,
		DestinationRepo: destRepo,
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

	destContext := destinationsContextJSON(dests)
	systemInstruction := fmt.Sprintf(`You are a warm, highly knowledgeable, and deeply hospitable local guide from Yogyakarta, Indonesia.
Your task is to act as a "knowledgeable local friend" helping tourists discover destinations in Yogyakarta.
Adopt a premium, elegant, yet warm and conversational tone.
Occasionally use gentle Javanese greetings (like 'Sugeng rawuh' for Welcome, 'Matur nuwun' for Thank you, 'Monggo' for Please proceed).
Answer inquiries thoroughly and recommend specific places from the list of actual destinations provided.

Here is the exact catalog of Yogyakarta destinations you can recommend. Do NOT invent new places; map the user's request intelligently to these options:
%s

Respond ONLY with valid JSON matching this schema:
{
  "reply": "Your friendly narrative advice, 3-5 sentences.",
  "matchedDestinationIds": ["array of destination IDs from the catalog"]
}`, destContext)

	userPrompt := buildUserPrompt(req.Query, req.History)

	result, err := h.AIService.Generate(context.Background(), ai.GenerateInput{
		SystemPrompt: systemInstruction,
		UserPrompt:   userPrompt,
		Temperature:  0.7,
		MaxTokens:    1500,
	})
	if err != nil {
		httpx.ErrorWithCode(c, 502, "AI_PROVIDER_ERROR", "Failed to generate recommendation")
		return
	}

	var parsed AIQueryResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		httpx.ErrorWithCode(c, 502, "AI_PROVIDER_ERROR", "Invalid AI response format")
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
		httpx.ErrorWithCode(c, 502, "AI_PROVIDER_ERROR", "Failed to analyze image")
		return
	}

	var parsed AIQueryResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		httpx.ErrorWithCode(c, 502, "AI_PROVIDER_ERROR", "Invalid AI response format")
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
		httpx.ErrorWithCode(c, 502, "AI_PROVIDER_ERROR", fmt.Sprintf("Failed to generate recommendation: %v", err))
		return
	}

	var parsed AIRecommendResponse
	if err := json.Unmarshal([]byte(result.Text), &parsed); err != nil {
		// fallback to offline if AI response is not parseable
		httpx.Success(c, 200, "Recommendation generated", h.offlineRecommendResponse(now), nil)
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
		ID        string `json:"id"`
		Name      string `json:"name"`
		Tagline   string `json:"tagline"`
		Category  string `json:"category"`
		BestTime  string `json:"bestTime"`
		SubRegion string `json:"subRegion"`
	}
	limit := len(dests)
	if limit > 25 {
		limit = 25
	}
	summaries := make([]destSummary, limit)
	for i := 0; i < limit; i++ {
		d := dests[i]
		summaries[i] = destSummary{
			ID: d.ExternalID, Name: d.Name, Tagline: d.Tagline,
			Category: d.Category, BestTime: d.BestTime, SubRegion: d.SubRegion,
		}
	}
	b, _ := json.Marshal(summaries) // compact JSON, no indent
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
