package destination

import (
	"fmt"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

// flexibleNumber accepts a JSON number or a numeric string and normalizes it
// to float64. An empty string or null leaves set=false so callers can treat it
// as "no change" instead of overwriting the stored value.
type flexibleNumber struct {
	value float64
	set   bool
}

func (f *flexibleNumber) UnmarshalJSON(b []byte) error {
	raw := strings.TrimSpace(strings.Trim(string(b), `"`))
	if raw == "" || raw == "null" {
		f.set = false
		return nil
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return err
	}
	f.value = v
	f.set = true
	return nil
}

func (s *Service) GetAll(status string) ([]Destination, error) {
	return s.Repo.FindAll(status)
}

func (s *Service) GetByID(externalID string) (*Destination, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetByCategory(category string) ([]Destination, error) {
	return s.Repo.FindByCategory(category)
}

func (s *Service) GetByRegion(region string, status string) ([]Destination, error) {
	return s.Repo.FindByRegion(region, status)
}

func (s *Service) Search(query string) ([]Destination, error) {
	return s.Repo.Search(query)
}

func (s *Service) UpdateUserDestination(userID uint, slug string, status string) error {
	return s.Repo.CreateOrUpdateUserDestination(userID, slug, status)
}

func (s *Service) GetUserDestinations(userID uint) ([]UserDestination, error) {
	return s.Repo.GetUserDestinations(userID)
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}

// CreateDestinationRequest is the payload accepted by POST /destinations.
type CreateDestinationRequest struct {
	Name              string         `json:"name"`
	Tagline           string         `json:"tagline"`
	Category          string         `json:"category"`
	Location          string         `json:"location"`
	SubRegion         string         `json:"sub_region"`
	Description       string         `json:"description"`
	Story             string         `json:"story"`
	TicketPrice       string         `json:"ticket_price"`
	OpeningHours      string         `json:"opening_hours"`
	BestTime          string         `json:"best_time"`
	Latitude          string         `json:"latitude"`
	Longitude         string         `json:"longitude"`
	Images            JSONArr        `json:"images"`
	Facilities        JSONArr        `json:"facilities"`
	TravelTips        JSONArr        `json:"travel_tips"`
	VideoUrl          string         `json:"video_url"`
	GoogleMapsURL     string         `json:"google_maps_url"`
	Rating            flexibleNumber `json:"rating"`
	ReviewCount       flexibleNumber `json:"review_count"`
	GoogleReviewCount flexibleNumber `json:"google_review_count"`
	SeoTitle          string         `json:"seo_title"`
	SeoKeywords       string         `json:"seo_keywords"`
	SeoDescription    string         `json:"seo_description"`
	Status            string         `json:"status"`
	// English translations
	NameEn           string  `json:"name_en"`
	TaglineEn        string  `json:"tagline_en"`
	DescriptionEn    string  `json:"description_en"`
	StoryEn          string  `json:"story_en"`
	BestTimeEn       string  `json:"best_time_en"`
	FacilitiesEn     JSONArr `json:"facilities_en"`
	TravelTipsEn     JSONArr `json:"travel_tips_en"`
	SeoTitleEn       string  `json:"seo_title_en"`
	SeoKeywordsEn    string  `json:"seo_keywords_en"`
	SeoDescriptionEn string  `json:"seo_description_en"`
}

// Create builds a Destination from the request, generates a slug-based
// ExternalID when none is provided, and persists it.
func (s *Service) Create(req CreateDestinationRequest) (*Destination, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	externalID := slug.Make(name)

	// Ensure the generated external_id is unique by appending a suffix if needed.
	baseID := externalID
	suffix := 1
	for {
		existing, err := s.Repo.FindByID(externalID)
		if err != nil || existing == nil {
			break
		}
		suffix++
		externalID = baseID + "-" + strconv.Itoa(suffix)
	}

	dest := Destination{
		ExternalID:       externalID,
		Name:             name,
		Tagline:          req.Tagline,
		Category:         req.Category,
		Location:         req.Location,
		SubRegion:        req.SubRegion,
		Description:      req.Description,
		Story:            req.Story,
		TicketPrice:      req.TicketPrice,
		OpeningHours:     req.OpeningHours,
		BestTime:         req.BestTime,
		Images:           req.Images,
		Facilities:       req.Facilities,
		TravelTips:       req.TravelTips,
		VideoURL:         req.VideoUrl,
		GoogleMapsURL:    req.GoogleMapsURL,
		SeoTitle:         req.SeoTitle,
		SeoKeywords:      req.SeoKeywords,
		SeoDescription:   req.SeoDescription,
		NameEn:           req.NameEn,
		TaglineEn:        req.TaglineEn,
		DescriptionEn:    req.DescriptionEn,
		StoryEn:          req.StoryEn,
		BestTimeEn:       req.BestTimeEn,
		FacilitiesEn:     req.FacilitiesEn,
		TravelTipsEn:     req.TravelTipsEn,
		SeoTitleEn:       req.SeoTitleEn,
		SeoKeywordsEn:    req.SeoKeywordsEn,
		SeoDescriptionEn: req.SeoDescriptionEn,
	}

	if req.Latitude != "" {
		if v, err := strconv.ParseFloat(req.Latitude, 64); err == nil {
			dest.Latitude = v
		}
	}
	if req.Longitude != "" {
		if v, err := strconv.ParseFloat(req.Longitude, 64); err == nil {
			dest.Longitude = v
		}
	}

	// Default status to "draft" when not provided.
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}
	dest.Status = status

	if req.Rating.set {
		dest.Rating = req.Rating.value
	}
	if req.ReviewCount.set {
		dest.ReviewCount = int(req.ReviewCount.value)
	}
	if req.GoogleReviewCount.set {
		dest.GoogleReviewCount = int(req.GoogleReviewCount.value)
	}

	if err := s.Repo.Create(&dest); err != nil {
		return nil, err
	}
	return &dest, nil
}

type UpdateDestinationRequest struct {
	Name              *string         `json:"name"`
	Tagline           *string         `json:"tagline"`
	Category          *string         `json:"category"`
	Location          *string         `json:"location"`
	SubRegion         *string         `json:"sub_region"`
	Description       *string         `json:"description"`
	Story             *string         `json:"story"`
	TicketPrice       *string         `json:"ticket_price"`
	OpeningHours      *string         `json:"opening_hours"`
	BestTime          *string         `json:"best_time"`
	Latitude          *string         `json:"latitude"`
	Longitude         *string         `json:"longitude"`
	Images            *JSONArr        `json:"images"`
	Facilities        *JSONArr        `json:"facilities"`
	TravelTips        *JSONArr        `json:"travel_tips"`
	VideoUrl          *string         `json:"video_url"`
	GoogleMapsURL     *string         `json:"google_maps_url"`
	FAQs              *JSONArr        `json:"faqs"`
	Weather           *JSONMap        `json:"weather"`
	Rating            *flexibleNumber `json:"rating"`
	ReviewCount       *flexibleNumber `json:"review_count"`
	GoogleReviewCount *flexibleNumber `json:"google_review_count"`
	SeoTitle          *string         `json:"seo_title"`
	SeoKeywords       *string         `json:"seo_keywords"`
	SeoDescription    *string         `json:"seo_description"`
	OgImageUrl        *string         `json:"og_image_url"`
	Status            *string         `json:"status"`
	HiddenGemOverride *string         `json:"hidden_gem_override"`
	// English translations
	NameEn           *string  `json:"name_en"`
	TaglineEn        *string  `json:"tagline_en"`
	DescriptionEn    *string  `json:"description_en"`
	StoryEn          *string  `json:"story_en"`
	BestTimeEn       *string  `json:"best_time_en"`
	FacilitiesEn     *JSONArr `json:"facilities_en"`
	TravelTipsEn     *JSONArr `json:"travel_tips_en"`
	SeoTitleEn       *string  `json:"seo_title_en"`
	SeoKeywordsEn    *string  `json:"seo_keywords_en"`
	SeoDescriptionEn *string  `json:"seo_description_en"`
}

func (s *Service) Update(externalID string, req UpdateDestinationRequest) (*Destination, error) {
	dest, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("destination not found")
	}

	if req.Name != nil {
		dest.Name = *req.Name
	}
	if req.Tagline != nil {
		dest.Tagline = *req.Tagline
	}
	if req.Category != nil {
		dest.Category = *req.Category
	}
	if req.Location != nil {
		dest.Location = *req.Location
	}
	if req.SubRegion != nil {
		dest.SubRegion = *req.SubRegion
	}
	if req.Description != nil {
		dest.Description = *req.Description
	}
	if req.Story != nil {
		dest.Story = *req.Story
	}
	if req.TicketPrice != nil {
		dest.TicketPrice = *req.TicketPrice
	}
	if req.OpeningHours != nil {
		dest.OpeningHours = *req.OpeningHours
	}
	if req.BestTime != nil {
		dest.BestTime = *req.BestTime
	}
	if req.Latitude != nil {
		if v, err := strconv.ParseFloat(*req.Latitude, 64); err == nil {
			dest.Latitude = v
		}
	}
	if req.Longitude != nil {
		if v, err := strconv.ParseFloat(*req.Longitude, 64); err == nil {
			dest.Longitude = v
		}
	}
	if req.Images != nil {
		dest.Images = *req.Images
	}
	if req.Facilities != nil {
		dest.Facilities = *req.Facilities
	}
	if req.TravelTips != nil {
		dest.TravelTips = *req.TravelTips
	}
	if req.VideoUrl != nil {
		dest.VideoURL = *req.VideoUrl
	}
	if req.GoogleMapsURL != nil {
		dest.GoogleMapsURL = *req.GoogleMapsURL
	}
	if req.FAQs != nil {
		dest.FAQs = *req.FAQs
	}
	if req.Weather != nil {
		dest.Weather = *req.Weather
	}
	if req.Rating != nil && req.Rating.set {
		dest.Rating = req.Rating.value
	}
	if req.ReviewCount != nil && req.ReviewCount.set {
		dest.ReviewCount = int(req.ReviewCount.value)
	}
	if req.GoogleReviewCount != nil && req.GoogleReviewCount.set {
		dest.GoogleReviewCount = int(req.GoogleReviewCount.value)
	}
	if req.SeoTitle != nil {
		dest.SeoTitle = *req.SeoTitle
	}
	if req.SeoKeywords != nil {
		dest.SeoKeywords = *req.SeoKeywords
	}
	if req.SeoDescription != nil {
		dest.SeoDescription = *req.SeoDescription
	}
	if req.OgImageUrl != nil {
		dest.OgImageUrl = *req.OgImageUrl
	}
	if req.Status != nil {
		// Quality gate: block direct publish if score below threshold
		if *req.Status == "published" {
			qs := CalculateScore(dest)
			if qs.Total < PublishScoreGate {
				return nil, fmt.Errorf(
					"quality gate: score %d/100 below minimum %d to publish (verdict: %s)",
					qs.Total, PublishScoreGate, qs.Verdict,
				)
			}
		}
		dest.Status = *req.Status
	}
	if req.HiddenGemOverride != nil {
		val := strings.TrimSpace(*req.HiddenGemOverride)
		// Validate allowed values.
		if val != "" && val != "pin" && val != "exclude" {
			return nil, fmt.Errorf("hidden_gem_override must be '', 'pin', or 'exclude'")
		}
		// Set pinned timestamp when admin pins for the first time or re-pins.
		if val == "pin" && dest.HiddenGemOverride != "pin" {
			now := time.Now()
			dest.HiddenGemPinnedAt = &now
		}
		// Clear timestamp when unpinning.
		if val != "pin" {
			dest.HiddenGemPinnedAt = nil
		}
		dest.HiddenGemOverride = val
	}
	if req.NameEn != nil {
		dest.NameEn = *req.NameEn
	}
	if req.TaglineEn != nil {
		dest.TaglineEn = *req.TaglineEn
	}
	if req.DescriptionEn != nil {
		dest.DescriptionEn = *req.DescriptionEn
	}
	if req.StoryEn != nil {
		dest.StoryEn = *req.StoryEn
	}
	if req.BestTimeEn != nil {
		dest.BestTimeEn = *req.BestTimeEn
	}
	if req.FacilitiesEn != nil {
		dest.FacilitiesEn = *req.FacilitiesEn
	}
	if req.TravelTipsEn != nil {
		dest.TravelTipsEn = *req.TravelTipsEn
	}
	if req.SeoTitleEn != nil {
		dest.SeoTitleEn = *req.SeoTitleEn
	}
	if req.SeoKeywordsEn != nil {
		dest.SeoKeywordsEn = *req.SeoKeywordsEn
	}
	if req.SeoDescriptionEn != nil {
		dest.SeoDescriptionEn = *req.SeoDescriptionEn
	}

	// Recalculate quality score before saving
	ApplyScoreToDestination(dest)

	if err := s.Repo.Update(dest); err != nil {
		return nil, err
	}
	return dest, nil
}

// recalcScore recalculates and persists the quality score for a destination.
// Called after any content-touching update.
func recalcScore(dest *Destination, repo Repository) {
	ApplyScoreToDestination(dest)
	// best-effort — don't fail the main operation if score update fails
	_ = repo.Update(dest)
}
