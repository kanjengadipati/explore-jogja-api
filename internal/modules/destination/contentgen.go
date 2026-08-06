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

// YouTubeFetcher is a function variable that allows injecting the YouTube fetch
// implementation without creating an import cycle (scraper → destination → scraper).
// Populated at startup by appsetup/router.go after both packages are initialized.
var YouTubeFetcher func(query string) string

// ─── Content status constants ─────────────────────────────────────────────────

const (
	ContentStatusDraft     = "draft"
	ContentStatusReview    = "review"
	ContentStatusPublished = "published"
	ContentStatusFlagged   = "flagged"
	ContentStatusRejected  = "rejected"
)

const (
	TemplateNarrative      = "narrative"
	TemplateFacts          = "facts"
	TemplateItineraryFirst = "itinerary_first"
)

// ─── Fact-density gate ────────────────────────────────────────────────────────

// factDensityScore counts how many key factual fields are populated.
// Canonical 10-field version — also exposed in DestinationResponse as
// fact_density_score so the frontend never needs to duplicate this logic.
func factDensityScore(d *Destination) int {
	score := 0
	if strings.TrimSpace(d.TicketPrice) != ""  { score++ }
	if strings.TrimSpace(d.OpeningHours) != "" { score++ }
	if len(d.Facilities) > 0                   { score++ }
	if len(d.TravelTips) > 0                   { score++ }
	if strings.TrimSpace(d.BestTime) != ""     { score++ }
	if d.Latitude != 0 && d.Longitude != 0    { score++ }
	if d.Rating > 0                            { score++ }
	if d.ReviewCount > 0                       { score++ }
	if strings.TrimSpace(d.Description) != "" { score++ }
	if strings.TrimSpace(d.Location) != ""    { score++ }
	return score
}

// ─── pg_trgm similarity gate ─────────────────────────────────────────────────

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

func (r *GormContentGenRepository) FindSimilarDescription(externalID, description string, threshold float64) ([]Destination, error) {
	var dests []Destination
	err := r.db.Raw(`
		SELECT * FROM destinations
		WHERE external_id != ?
		  AND content_status = 'published'
		  AND word_similarity(description, ?) > ?
		LIMIT 5
	`, externalID, description, threshold).Scan(&dests).Error
	if err != nil && strings.Contains(err.Error(), "function word_similarity") {
		return []Destination{}, nil
	}
	return dests, err
}

// ─── ContentGen service ───────────────────────────────────────────────────────

type ContentGenService struct {
	DestRepo            Repository
	ContentRepo         ContentGenRepository
	AIService           *ai.Service
	SimilarityThreshold float64
}

func NewContentGenService(destRepo Repository, contentRepo ContentGenRepository, aiService *ai.Service) *ContentGenService {
	return &ContentGenService{
		DestRepo:            destRepo,
		ContentRepo:         contentRepo,
		AIService:           aiService,
		SimilarityThreshold: 0.6,
	}
}

func (s *ContentGenService) Generate(ctx context.Context, externalID string, variant string) (*Destination, error) {
	dest, err := s.DestRepo.FindByID(externalID)
	if err != nil {
		return nil, fmt.Errorf("destination not found: %w", err)
	}

	if variant == "" {
		variant = TemplateNarrative
	}

	prompt := buildPrompt(dest, variant)
	result, err := s.AIService.Generate(ctx, ai.GenerateInput{
		SystemPrompt: prompt.system,
		UserPrompt:   prompt.user,
		Temperature:  0.7,
		MaxTokens:    2500,
	})
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	var generated map[string]interface{}
	if err := json.Unmarshal([]byte(result.Text), &generated); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	applyGeneratedContent(dest, generated)

	// Auto-fill video_url via YouTube Data API if empty and YouTubeFetcher is wired.
	// Fail-open — if API key is absent or quota exhausted, generation still succeeds.
	if dest.VideoURL == "" && YouTubeFetcher != nil {
		searchQuery := dest.Name
		if dest.Location != "" {
			searchQuery += " " + dest.Location
		}
		searchQuery += " wisata yogyakarta"
		if videoURL := YouTubeFetcher(searchQuery); videoURL != "" {
			dest.VideoURL = videoURL
		}
	}

	dest.ContentStatus = ContentStatusDraft
	dest.TemplateVariant = variant

	ApplyScoreToDestination(dest)

	if err := s.DestRepo.Update(dest); err != nil {
		return nil, fmt.Errorf("failed to save draft: %w", err)
	}
	return dest, nil
}

func (s *ContentGenService) Approve(ctx context.Context, externalID string) (*Destination, error) {
	dest, err := s.DestRepo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("destination not found")
	}

	ApplyScoreToDestination(dest)

	if dest.ContentScore < PublishScoreGate {
		return nil, fmt.Errorf(
			"quality gate failed: score %d/100 is below minimum %d required to publish (verdict: %s)",
			dest.ContentScore, PublishScoreGate, dest.ContentVerdict,
		)
	}

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

	schema := `Return ONLY valid JSON with these fields:
description (string, Indonesian, 2-3 paragraphs),
description_en (string, English),
story (string, Indonesian, editorial narrative ≥300 chars),
story_en (string, English),
tagline (string, Indonesian, 5-8 words),
tagline_en (string, English),
seo_title (string, Indonesian, max 60 chars),
seo_title_en (string, English, max 60 chars),
seo_description (string, Indonesian, max 160 chars),
seo_description_en (string, English, max 160 chars),
seo_keywords (string, Indonesian, comma-separated),
seo_keywords_en (string, English, comma-separated),
best_time (string, Indonesian),
best_time_en (string, English),
facilities (array of strings, Indonesian, min 3 items — physical amenities present at the site),
facilities_en (array of strings, English),
travel_tips (array of strings, Indonesian, min 3 practical visitor tips),
travel_tips_en (array of strings, English).`

	var systemInstruction, userPrompt string
	switch variant {
	case TemplateFacts:
		systemInstruction = base + "\nSTYLE: Facts-first. Lead with concrete numbers, prices, facilities, and practical info.\n" + schema
		userPrompt = fmt.Sprintf("Write facts-first content for %s. Prioritize practical info, facilities, and travel tips.", d.Name)
	case TemplateItineraryFirst:
		systemInstruction = base + "\nSTYLE: Itinerary-first. Open with a narrative day plan. Then supporting facts and facilities.\n" + schema
		userPrompt = fmt.Sprintf("Write itinerary-first content for %s. Lead with how to spend time there, then list facilities and tips.", d.Name)
	default:
		systemInstruction = base + "\nSTYLE: Narrative. Open with atmosphere and emotion. Facts and facilities follow the story.\n" + schema
		userPrompt = fmt.Sprintf("Write narrative content for %s. Open with the feeling of being there, then practical details.", d.Name)
	}

	return builtPrompt{system: systemInstruction, user: userPrompt}
}

// applyGeneratedContent merges AI-generated fields into the destination struct.
// Only non-empty values are applied — existing data is never blanked by AI output.
func applyGeneratedContent(d *Destination, generated map[string]interface{}) {
	setStr := func(field *string, key string) {
		if v, ok := generated[key].(string); ok && strings.TrimSpace(v) != "" {
			*field = v
		}
	}
	setArr := func(field *JSONArr, key string) {
		v, ok := generated[key].([]interface{})
		if !ok || len(v) == 0 {
			return
		}
		arr := make(JSONArr, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				arr = append(arr, s)
			}
		}
		if len(arr) > 0 {
			*field = arr
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
	setStr(&d.BestTime, "best_time")
	setStr(&d.BestTimeEn, "best_time_en")

	setArr(&d.Facilities, "facilities")
	setArr(&d.FacilitiesEn, "facilities_en")
	setArr(&d.TravelTips, "travel_tips")
	setArr(&d.TravelTipsEn, "travel_tips_en")

	// Auto-fill GoogleMapsURL from coordinates if not already set
	AutoFillGoogleMapsURL(d)
}

// ─── HTTP handler methods ─────────────────────────────────────────────────────

type ContentGenHandler struct {
	Service *ContentGenService
}

func NewContentGenHandler(svc *ContentGenService) *ContentGenHandler {
	return &ContentGenHandler{Service: svc}
}

type contentQueueItem struct {
	Destination
	FactDensityScore int `json:"fact_density_score"`
}

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

func (h *ContentGenHandler) GenerateDraft(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Variant string `json:"variant"`
	}
	_ = c.ShouldBindJSON(&req)
	dest, err := h.Service.Generate(c.Request.Context(), id, req.Variant)
	if err != nil {
		httpx.ErrorWithCode(c, 400, "GENERATION_FAILED", err.Error())
		return
	}
	httpx.Success(c, 200, "Draft generated", dest, nil)
}

func (h *ContentGenHandler) ApproveDraft(c *gin.Context) {
	id := c.Param("id")
	dest, err := h.Service.Approve(c.Request.Context(), id)
	if err != nil {
		httpx.ErrorWithCode(c, 409, "APPROVAL_FAILED", err.Error())
		return
	}
	httpx.Success(c, 200, "Content published", dest, nil)
}

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
