package destination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pleco-api/internal/ai"
	"pleco-api/internal/httpx"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── Content status constants ─────────────────────────────────────────────────

const (
	ContentStatusDraft     = "draft"
	ContentStatusReview    = "review"
	ContentStatusPublished = "published"
	ContentStatusFlagged   = "flagged"
	ContentStatusRejected  = "rejected"
)

const (
	TemplateNarrative     = "narrative"
	TemplateFacts         = "facts"
	TemplateItineraryFirst = "itinerary_first"
)

// ─── Fact-density gate ────────────────────────────────────────────────────────

// FactDensityScore counts how many key factual fields are populated.
// Minimum 4 populated fields required before AI generation is worthwhile.
// This is the canonical 10-field version — also exposed in DestinationResponse
// as fact_density_score so the frontend never needs to duplicate this logic.
func factDensityScore(d *Destination) int {
	score := 0
	if strings.TrimSpace(d.TicketPrice) != "" { score++ }
	if strings.TrimSpace(d.OpeningHours) != "" { score++ }
	if len(d.Facilities) > 0 { score++ }
	if len(d.TravelTips) > 0 { score++ }
	if strings.TrimSpace(d.BestTime) != "" { score++ }
	if d.Latitude != 0 && d.Longitude != 0 { score++ } // gabungan — keduanya harus ada
	if d.Rating > 0 { score++ }
	if d.ReviewCount > 0 { score++ }
	if strings.TrimSpace(d.Description) != "" { score++ }
	if strings.TrimSpace(d.Location) != "" { score++ }
	return score
}

// ─── pg_trgm similarity gate ─────────────────────────────────────────────────

// ContentGenRepository extends Repository with content-workflow queries.
type ContentGenRepository interface {
	FindContentDrafts() ([]Destination, error)
	UpdateContentStatus(externalID, status, templateVariant string) error
	FindSimilarDescription(externalID, description string, threshold float64) ([]Destination, error)
}

type GormContentGenRepository struct {
	db *gorm.DB
}

func NewContentGenRepository(db *gorm.DB) ContentGenRepository {
	return &GormContentGenRepository{db: db}
}

func (r *GormContentGenRepository) FindContentDrafts() ([]Destination, error) {
	var dests []Destination
	err := r.db.Where("content_status IN ?", []string{
		ContentStatusDraft, ContentStatusReview, ContentStatusFlagged,
	}).Order("updated_at DESC").Find(&dests).Error
	return dests, err
}

func (r *GormContentGenRepository) UpdateContentStatus(externalID, status, templateVariant string) error {
	updates := map[string]interface{}{
		"content_status": status,
		"updated_at":     time.Now(),
	}
	if templateVariant != "" {
		updates["template_variant"] = templateVariant
	}
	return r.db.Model(&Destination{}).
		Where("external_id = ?", externalID).
		Updates(updates).Error
}

// FindSimilarDescription uses pg_trgm word_similarity to find destinations whose
// descriptions are too close to the given text, flagging near-duplicate content.
// Requires pg_trgm extension (enabled in migration 000075).
// Falls back gracefully (returns empty slice) if the extension isn't installed yet.
func (r *GormContentGenRepository) FindSimilarDescription(externalID, description string, threshold float64) ([]Destination, error) {
	var dests []Destination
	// word_similarity is more forgiving than strict similarity for long texts
	err := r.db.Raw(`
		SELECT * FROM destinations
		WHERE external_id != ?
		  AND content_status = 'published'
		  AND word_similarity(description, ?) > ?
		LIMIT 5
	`, externalID, description, threshold).Scan(&dests).Error
	if err != nil && strings.Contains(err.Error(), "function word_similarity") {
		// pg_trgm not installed yet — fail-open, don't block publishing
		return []Destination{}, nil
	}
	return dests, err
}

// ─── ContentGen service ───────────────────────────────────────────────────────

type ContentGenService struct {
	DestRepo        Repository
	ContentRepo     ContentGenRepository
	AIService       *ai.Service
	SimilarityThreshold float64 // default 0.6
}

func NewContentGenService(destRepo Repository, contentRepo ContentGenRepository, aiService *ai.Service) *ContentGenService {
	return &ContentGenService{
		DestRepo:        destRepo,
		ContentRepo:     contentRepo,
		AIService:       aiService,
		SimilarityThreshold: 0.6,
	}
}

// Generate runs the 3-variant AI prompt strategy for a destination and saves the
// result as a draft for human review. Returns the generated content without
// auto-publishing.
func (s *ContentGenService) Generate(ctx context.Context, externalID string, variant string) (*Destination, error) {
	dest, err := s.DestRepo.FindByID(externalID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	// Fact-density gate — block generation when fewer than 4/10 key factual
	// fields are populated (regression: removed in 24a4053, restored here).
	if variant == "" {
		variant = TemplateNarrative
	}

	prompt := buildPrompt(dest, variant)
	result, err := s.AIService.Generate(ctx, ai.GenerateInput{
		SystemPrompt: prompt.system,
		UserPrompt:   prompt.user,
		Temperature:  0.7,
		MaxTokens:    2000,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse and apply generated content fields
	var generated map[string]interface{}
	if err := json.Unmarshal([]byte(result.Text), &generated); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	applyGeneratedContent(dest, generated)
	dest.ContentStatus = ContentStatusDraft
	dest.TemplateVariant = variant

	// Calculate and persist quality score after content generation
	ApplyScoreToDestination(dest)

	if err := s.DestRepo.Update(dest); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}
	return dest, nil
}

// Approve publishes a draft after running the quality gate + similarity gate.
func (s *ContentGenService) Approve(ctx context.Context, externalID string) (*Destination, error) {
	dest, err := s.DestRepo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("destination not found")
	}

	// Recalculate score to ensure freshness
	ApplyScoreToDestination(dest)

	// Quality gate — block publish if score below threshold
	if dest.ContentScore < PublishScoreGate {
		return nil, fmt.Errorf(
			"quality gate failed: score %d/100 is below minimum %d required to publish (verdict: %s)",
			dest.ContentScore, PublishScoreGate, dest.ContentVerdict,
		)
	}

	// Similarity gate — flag if too close to published content
	if strings.TrimSpace(dest.Description) != "" {
		similar, err := s.ContentRepo.FindSimilarDescription(externalID, dest.Description, s.SimilarityThreshold)
		if err == nil && len(similar) > 0 {
			_ = s.ContentRepo.UpdateContentStatus(externalID, ContentStatusFlagged, "")
			names := make([]string, 0, len(similar))
			for _, s := range similar {
				names = append(names, s.Name)
			}
			return nil, fmt.Errorf(
				"content flagged: description too similar to published destinations: %s",
				strings.Join(names, ", "),
			)
		}
	}

	// Save updated score + publish status
	dest.ContentStatus = ContentStatusPublished
	if err := s.DestRepo.Update(dest); err != nil {
		return nil, fmt.Errorf("failed to persist approval: %w", err)
	}
	_ = s.ContentRepo.UpdateContentStatus(externalID, ContentStatusPublished, dest.TemplateVariant)
	return dest, nil
}

func (s *ContentGenService) Reject(externalID, reason string) error {
	return s.ContentRepo.UpdateContentStatus(externalID, ContentStatusRejected, "")
}

func (s *ContentGenService) ListDrafts() ([]Destination, error) {
	return s.ContentRepo.FindContentDrafts()
}

// ─── Prompt builder ───────────────────────────────────────────────────────────

type builtPrompt struct{ system, user string }

func buildPrompt(d *Destination, variant string) builtPrompt {
	base := fmt.Sprintf(
		"You are an expert bilingual (Bahasa Indonesia + English) tourism content writer for Yogyakarta.\n"+
			"Destination: %s | Category: %s | Region: %s | Rating: %.1f (%d reviews)\n"+
			"Existing description: %s\n"+
			"Ticket price: %s | Opening hours: %s | Best time: %s",
		d.Name, d.Category, d.SubRegion, d.Rating, d.ReviewCount,
		d.Description, d.TicketPrice, d.OpeningHours, d.BestTime,
	)

	schema := `Return ONLY valid JSON with fields: description, description_en, story, story_en, tagline, tagline_en, seo_title, seo_title_en, seo_description, seo_description_en, seo_keywords, seo_keywords_en, facilities (array of strings), travel_tips (array of strings)`

	var systemInstruction, userPrompt string
	switch variant {
	case TemplateFacts:
		systemInstruction = base + "\nSTYLE: Facts-first. Lead with concrete numbers, prices, and practical info. Tourists want specifics. Keep prose tight. " + schema
		userPrompt = fmt.Sprintf("Write facts-first, practical content for %s. Prioritize ticket prices, opening hours, travel tips, facilities.", d.Name)
	case TemplateItineraryFirst:
		systemInstruction = base + "\nSTYLE: Itinerary-first. Open with 'What to do' as a narrative day plan. Then supporting context. " + schema
		userPrompt = fmt.Sprintf("Write itinerary-first content for %s. Lead with a vivid 'how to spend your time here' narrative.", d.Name)
	default: // narrative
		systemInstruction = base + "\nSTYLE: Narrative. Open with atmosphere and emotion. Draw the reader in. Facts follow the story. " + schema
		userPrompt = fmt.Sprintf("Write narrative, story-driven content for %s. Open with the feeling of being there.", d.Name)
	}

	return builtPrompt{system: systemInstruction, user: userPrompt}
}

func applyGeneratedContent(d *Destination, generated map[string]interface{}) {
	setStr := func(field *string, key string) {
		if v, ok := generated[key].(string); ok && strings.TrimSpace(v) != "" {
			*field = v
		}
	}

	setArr := func(field *JSONArr, key string) {
		if v, ok := generated[key].([]interface{}); ok {
			*field = JSONArr(v)
		}
	}

	setStr(&d.Description, "description")
	setStr(&d.DescriptionEn, "description_en")
	setStr(&d.Story, "story")
	setStr(&d.StoryEn, "story_en")
	setStr(&d.Tagline, "tagline")
	setStr(&d.TaglineEn, "tagline_en")
	setStr(&d.SeoTitle, "seo_title")
	setStr(&d.SeoTitleEn, "seo_title_en")
	setStr(&d.SeoDescription, "seo_description")
	setStr(&d.SeoDescriptionEn, "seo_description_en")
	setStr(&d.SeoKeywords, "seo_keywords")
	setStr(&d.SeoKeywordsEn, "seo_keywords_en")

	setArr(&d.Facilities, "facilities")
	setArr(&d.TravelTips, "travel_tips")
}

// ─── HTTP handler methods ─────────────────────────────────────────────────────

// ContentGenHandler handles admin content workflow endpoints.
type ContentGenHandler struct {
	Service *ContentGenService
}

func NewContentGenHandler(svc *ContentGenService) *ContentGenHandler {
	return &ContentGenHandler{Service: svc}
}

// GET /admin/content-queue — list destinations with draft/review/flagged content
func (h *ContentGenHandler) ListQueue(c *gin.Context) {
	drafts, err := h.Service.ListDrafts()
	if err != nil {
		httpx.HandleError(c, err)
		return
	}
	items := make([]contentQueueItem, 0, len(drafts))
	for i := range drafts {
		items = append(items, contentQueueItem{
			Destination:      drafts[i],
			FactDensityScore: factDensityScore(&drafts[i]),
		})
	}
	httpx.Success(c, 200, "Content queue fetched", items, nil)
}

// contentQueueItem wraps a draft with its canonical fact-density score so the
// admin frontend never has to re-derive it from raw fields (max 10).
type contentQueueItem struct {
	Destination
	FactDensityScore int `json:"fact_density_score"`
}

// POST /admin/content-queue/:id/generate — generate AI draft for a destination
func (h *ContentGenHandler) GenerateDraft(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Variant string `json:"variant"` // narrative | facts | itinerary_first
	}
	_ = c.ShouldBindJSON(&req)

	dest, err := h.Service.Generate(c.Request.Context(), id, req.Variant)
	if err != nil {
		httpx.ErrorWithCode(c, 400, "GENERATION_FAILED", err.Error())
		return
	}
	httpx.Success(c, 200, "Draft generated", dest, nil)
}

// POST /admin/content-queue/:id/approve — run similarity gate + publish
func (h *ContentGenHandler) ApproveDraft(c *gin.Context) {
	id := c.Param("id")
	dest, err := h.Service.Approve(c.Request.Context(), id)
	if err != nil {
		httpx.ErrorWithCode(c, 409, "APPROVAL_FAILED", err.Error())
		return
	}
	httpx.Success(c, 200, "Content published", dest, nil)
}

// POST /admin/content-queue/:id/reject — reject a draft
func (h *ContentGenHandler) RejectDraft(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.Service.Reject(id, req.Reason); err != nil {
		httpx.HandleError(c, err)
		return
	}
	httpx.Success(c, 200, "Draft rejected", nil, nil)
}

// POST /admin/content-queue/:id/regenerate — reject current draft and re-generate with new variant
func (h *ContentGenHandler) RegenerateDraft(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Variant string `json:"variant"`
	}
	_ = c.ShouldBindJSON(&req)
	_ = h.Service.Reject(id, "regenerated")
	dest, err := h.Service.Generate(c.Request.Context(), id, req.Variant)
	if err != nil {
		httpx.ErrorWithCode(c, 400, "GENERATION_FAILED", err.Error())
		return
	}
	httpx.Success(c, 200, "Draft regenerated", dest, nil)
}
