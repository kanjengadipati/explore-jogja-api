package destination

import "errors"

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]Destination, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*Destination, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetByCategory(category string) ([]Destination, error) {
	return s.Repo.FindByCategory(category)
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

type UpdateDestinationRequest struct {
	Name              *string   `json:"name"`
	Tagline           *string   `json:"tagline"`
	Category          *string   `json:"category"`
	Location          *string   `json:"location"`
	SubRegion         *string   `json:"sub_region"`
	Description       *string   `json:"description"`
	Story             *string   `json:"story"`
	TicketPrice       *string   `json:"ticket_price"`
	OpeningHours      *string   `json:"opening_hours"`
	BestTime          *string   `json:"best_time"`
	Latitude          *float64  `json:"latitude"`
	Longitude         *float64  `json:"longitude"`
	Images            *JSONArr  `json:"images"`
	Facilities        *JSONArr  `json:"facilities"`
	TravelTips        *JSONArr  `json:"travel_tips"`
	VideoUrl          *string   `json:"video_url"`
	GoogleMapsURL     *string   `json:"google_maps_url"`
	GoogleReviewCount *int      `json:"google_review_count"`
	SeoTitle          *string   `json:"seo_title"`
	SeoKeywords       *string   `json:"seo_keywords"`
	SeoDescription    *string   `json:"seo_description"`
	OgImageUrl        *string   `json:"og_image_url"`
	// English translations
	NameEn        *string  `json:"name_en"`
	TaglineEn     *string  `json:"tagline_en"`
	DescriptionEn *string  `json:"description_en"`
	StoryEn       *string  `json:"story_en"`
	BestTimeEn    *string  `json:"best_time_en"`
	FacilitiesEn  *JSONArr `json:"facilities_en"`
	TravelTipsEn  *JSONArr `json:"travel_tips_en"`
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
		dest.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		dest.Longitude = *req.Longitude
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
	if req.GoogleReviewCount != nil {
		dest.GoogleReviewCount = *req.GoogleReviewCount
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

	if err := s.Repo.Update(dest); err != nil {
		return nil, err
	}
	return dest, nil
}
