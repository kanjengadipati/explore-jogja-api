package adcampaign

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/subscription"
)

type Service struct {
	Repo           Repository
	SubscriptionSvc *subscription.Service
}

func NewService(repo Repository, subscriptionSvc *subscription.Service) *Service {
	return &Service{Repo: repo, SubscriptionSvc: subscriptionSvc}
}

func (s *Service) GetAll() ([]AdCampaign, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetAllByBusiness(businessExternalID string) ([]AdCampaign, error) {
	return s.Repo.FindAllByBusinessExternalID(businessExternalID)
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
	if campaign.BusinessExternalID == nil || *campaign.BusinessExternalID == "" {
		return errors.New("business_external_id is required")
	}
	exists, err := s.Repo.BusinessExists(*campaign.BusinessExternalID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("business not found")
	}

	canCreate, err := s.SubscriptionSvc.CanCreateAdCampaign(*campaign.BusinessExternalID)
	if err != nil {
		return err
	}
	if !canCreate {
		return errors.New("business on free plan cannot create ad campaigns")
	}

	return s.Repo.Create(campaign)
}

type UpdateAdCampaignRequest struct {
	PartnerName        *string    `json:"partner_name"`
	BusinessExternalID *string    `json:"business_external_id"`
	Placement          *string    `json:"placement"`
	ImageURL           *string    `json:"image_url"`
	TargetURL          *string    `json:"target_url"`
	Category           *string    `json:"category"`
	StartAt            *time.Time `json:"start_at"`
	EndAt              *time.Time `json:"end_at"`
	Weight             *int       `json:"weight"`
	IsActive           *bool      `json:"is_active"`

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
	if req.BusinessExternalID != nil {
		if *req.BusinessExternalID == "" {
			return nil, errors.New("business_external_id cannot be empty")
		}
		exists, err := s.Repo.BusinessExists(*req.BusinessExternalID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, errors.New("business not found")
		}
		campaign.BusinessExternalID = req.BusinessExternalID
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

func (s *Service) GetAllHouseAds() ([]HouseAd, error) {
	return s.Repo.FindAllHouseAds()
}

func (s *Service) GetEnabledHouseAd(placement string) (*HouseAd, error) {
	return s.Repo.FindEnabledHouseAdByPlacement(placement)
}

type UpdateHouseAdRequest struct {
	Placement *string `json:"placement"`
	Headline  *string `json:"headline"`
	Subline   *string `json:"subline"`
	CTALabel  *string `json:"cta_label"`
	ImageURL  *string `json:"image_url"`
	TargetURL *string `json:"target_url"`
	IsEnabled *bool   `json:"is_enabled"`
}

func (s *Service) CreateHouseAd(houseAd *HouseAd) error {
	if houseAd.ExternalID == "" {
		houseAd.ExternalID = uuid.New().String()
	}
	return s.Repo.CreateHouseAd(houseAd)
}

func (s *Service) UpdateHouseAd(externalID string, req UpdateHouseAdRequest) (*HouseAd, error) {
	houseAd, err := s.Repo.FindHouseAdByID(externalID)
	if err != nil {
		return nil, errors.New("house ad not found")
	}
	if req.Placement != nil {
		houseAd.Placement = *req.Placement
	}
	if req.Headline != nil {
		houseAd.Headline = *req.Headline
	}
	if req.Subline != nil {
		houseAd.Subline = *req.Subline
	}
	if req.CTALabel != nil {
		houseAd.CTALabel = *req.CTALabel
	}
	if req.ImageURL != nil {
		houseAd.ImageURL = *req.ImageURL
	}
	if req.TargetURL != nil {
		houseAd.TargetURL = *req.TargetURL
	}
	if req.IsEnabled != nil {
		houseAd.IsEnabled = *req.IsEnabled
	}
	if err := s.Repo.UpdateHouseAd(houseAd); err != nil {
		return nil, err
	}
	return houseAd, nil
}

func (s *Service) DeleteHouseAd(externalID string) error {
	return s.Repo.DeleteHouseAd(externalID)
}
