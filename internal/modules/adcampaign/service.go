package adcampaign

import (
	"errors"
	"time"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) GetAll() ([]AdCampaign, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByID(externalID string) (*AdCampaign, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetActiveBanner(placement, category string) (*AdCampaign, error) {
	candidates, err := s.Repo.FindActiveCandidates(placement, category)
	if err != nil {
		return nil, err
	}
	return WeightedPick(candidates), nil
}

func (s *Service) Create(campaign *AdCampaign) error {
	return s.Repo.Create(campaign)
}

type UpdateAdCampaignRequest struct {
	PartnerName *string    `json:"partner_name"`
	Placement   *string    `json:"placement"`
	ImageURL    *string    `json:"image_url"`
	TargetURL   *string    `json:"target_url"`
	Category    *string    `json:"category"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	Weight      *int       `json:"weight"`
	IsActive    *bool      `json:"is_active"`

	PriceAmount   *float64 `json:"price_amount"`
	PriceCurrency *string  `json:"price_currency"`
	PaymentStatus *string  `json:"payment_status"`
}

func (s *Service) Update(externalID string, req UpdateAdCampaignRequest) (*AdCampaign, error) {
	campaign, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, errors.New("ad campaign not found")
	}

	if req.PartnerName != nil {
		campaign.PartnerName = *req.PartnerName
	}
	if req.Placement != nil {
		campaign.Placement = *req.Placement
	}
	if req.ImageURL != nil {
		campaign.ImageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		campaign.TargetURL = *req.TargetURL
	}
	if req.Category != nil {
		campaign.Category = *req.Category
	}
	if req.StartAt != nil {
		campaign.StartAt = *req.StartAt
	}
	if req.EndAt != nil {
		campaign.EndAt = *req.EndAt
	}
	if req.Weight != nil {
		campaign.Weight = *req.Weight
	}
	if req.IsActive != nil {
		campaign.IsActive = *req.IsActive
	}
	if req.PriceAmount != nil {
		campaign.PriceAmount = *req.PriceAmount
	}
	if req.PriceCurrency != nil {
		campaign.PriceCurrency = *req.PriceCurrency
	}
	if req.PaymentStatus != nil {
		campaign.PaymentStatus = *req.PaymentStatus
	}

	if err := s.Repo.Update(campaign); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}

func (s *Service) TrackImpression(externalID string) error {
	return s.Repo.IncrementImpression(externalID)
}

func (s *Service) TrackClick(externalID string) error {
	return s.Repo.IncrementClick(externalID)
}
