package partner

import (
	"errors"
	"time"

	"pleco-api/internal/modules/notification"
	"pleco-api/internal/modules/user"

	"github.com/google/uuid"
)

type Service struct {
	Repo    Repository
	UserSvc *user.Service
	NotifSvc *notification.Service
}

func NewService(repo Repository, userSvc *user.Service, notifSvc *notification.Service) *Service {
	return &Service{Repo: repo, UserSvc: userSvc, NotifSvc: notifSvc}
}

// --- Public (approved only) ---

func (s *Service) GetAllApproved() ([]Partner, error) {
	return s.Repo.FindAllApproved()
}

func (s *Service) GetByID(externalID string) (*Partner, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) Search(query string) ([]Partner, error) {
	return s.Repo.Search(query)
}

func (s *Service) GetSponsored(destinationID, category string) ([]Partner, error) {
	candidates, err := s.Repo.FindSponsored(destinationID, category)
	if err != nil {
		return nil, err
	}
	// Fair weighted random selection by sponsor_tier (§3.5)
	picked := WeightedPickPartner(candidates)
	if picked == nil {
		return []Partner{}, nil
	}
	return []Partner{*picked}, nil
}

func (s *Service) TrackImpression(externalID string) error {
	_ = s.Repo.IncrementDailyStats(externalID, false)
	return s.Repo.IncrementImpression(externalID)
}

func (s *Service) TrackClick(externalID string) error {
	_ = s.Repo.IncrementDailyStats(externalID, true)
	return s.Repo.IncrementClick(externalID)
}

func (s *Service) GetDailyStats(externalID string, startDate, endDate time.Time) ([]PartnerDailyStats, error) {
	return s.Repo.FindDailyStats(externalID, startDate, endDate)
}

// --- Admin (any status) ---

func (s *Service) GetByIDAny(externalID string) (*Partner, error) {
	return s.Repo.FindByIDAny(externalID)
}

func (s *Service) GetAllAny() ([]Partner, error) {
	return s.Repo.FindAllAny()
}

func (s *Service) GetByStatus(status string) ([]Partner, error) {
	return s.Repo.FindByStatus(status)
}

func (s *Service) Create(partner *Partner) error {
	return s.Repo.Create(partner)
}

type AdminCreatePartnerRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Image       string  `json:"image"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Price       string  `json:"price"`
	OwnerUserID *uint   `json:"owner_user_id"`
	Status      string  `json:"status"`
}

func (s *Service) AdminCreate(req AdminCreatePartnerRequest) (*Partner, error) {
	partner := Partner{
		ExternalID:   uuid.New().String(),
		Name:         req.Name,
		Category:     req.Category,
		Description:  req.Description,
		Location:     req.Location,
		Address:      req.Address,
		Image:        req.Image,
		Phone:        req.Phone,
		Website:      req.Website,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		Price:        req.Price,
		OwnerUserID:  req.OwnerUserID,
		Status:       req.Status,
	}

	if err := s.Repo.Create(&partner); err != nil {
		return nil, err
	}
	return &partner, nil
}


func (s *Service) Save(partner *Partner) error {
	return s.Repo.Update(partner)
}

type UpdatePartnerRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Category    *string  `json:"category"`
	Location    *string  `json:"location"`
	Address     *string  `json:"address"`
	Image       *string  `json:"image"`
	Rating      *float64 `json:"rating"`
	Price       *string  `json:"price"`
	Distance    *string  `json:"distance"`
	Phone       *string  `json:"phone"`
	Website     *string  `json:"website"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`

	IsSponsored    *bool      `json:"is_sponsored"`
	SponsorTier    *int       `json:"sponsor_tier"`
	SponsorStartAt *time.Time `json:"sponsor_start_at"`
	SponsorEndAt   *time.Time `json:"sponsor_end_at"`
	TargetDestIDs  *JSONArr   `json:"target_dest_ids"`

	SponsorPrice         *float64 `json:"sponsor_price"`
	SponsorPriceCurrency *string  `json:"sponsor_price_currency"`
	SponsorPaymentStatus *string  `json:"sponsor_payment_status"`
}

func (s *Service) Update(externalID string, req UpdatePartnerRequest) (*Partner, error) {
	partner, err := s.Repo.FindByIDAny(externalID)
	if err != nil {
		return nil, errors.New("partner not found")
	}

	if req.Name != nil {
		partner.Name = *req.Name
	}
	if req.Description != nil {
		partner.Description = *req.Description
	}
	if req.Category != nil {
		partner.Category = *req.Category
	}
	if req.Location != nil {
		partner.Location = *req.Location
	}
	if req.Address != nil {
		partner.Address = *req.Address
	}
	if req.Image != nil {
		partner.Image = *req.Image
	}
	if req.Rating != nil {
		partner.Rating = *req.Rating
	}
	if req.Price != nil {
		partner.Price = *req.Price
	}
	if req.Distance != nil {
		partner.Distance = *req.Distance
	}
	if req.Phone != nil {
		partner.Phone = *req.Phone
	}
	if req.Website != nil {
		partner.Website = *req.Website
	}
	if req.Latitude != nil {
		partner.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		partner.Longitude = *req.Longitude
	}
	if req.IsSponsored != nil {
		partner.IsSponsored = *req.IsSponsored
	}
	if req.SponsorTier != nil {
		partner.SponsorTier = *req.SponsorTier
	}
	if req.SponsorStartAt != nil {
		partner.SponsorStartAt = *req.SponsorStartAt
	}
	if req.SponsorEndAt != nil {
		partner.SponsorEndAt = *req.SponsorEndAt
	}
	if req.TargetDestIDs != nil {
		partner.TargetDestIDs = *req.TargetDestIDs
	}
	if req.SponsorPrice != nil {
		partner.SponsorPrice = *req.SponsorPrice
	}
	if req.SponsorPriceCurrency != nil {
		partner.SponsorPriceCurrency = *req.SponsorPriceCurrency
	}
	if req.SponsorPaymentStatus != nil {
		partner.SponsorPaymentStatus = *req.SponsorPaymentStatus
	}

	if err := s.Repo.Update(partner); err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *Service) Delete(externalID string) error {
	return s.Repo.Delete(externalID)
}

// --- Self-service ---

type ApplyPartnerRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Image       string  `json:"image"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Price       string  `json:"price"`
}

func (s *Service) Apply(req ApplyPartnerRequest, ownerUserID uint) (*Partner, error) {
	now := time.Now()
	partner := Partner{
		ExternalID:  uuid.New().String(),
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Location:    req.Location,
		Address:     req.Address,
		Image:       req.Image,
		Phone:       req.Phone,
		Website:     req.Website,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Price:       req.Price,
		OwnerUserID: &ownerUserID,
		Status:      StatusPending,
		SubmittedAt: &now,
	}

	if err := s.Repo.Create(&partner); err != nil {
		return nil, err
	}

	// Upgrade caller's role to "partner" so they can access self-service endpoints.
	// Idempotent — no-op if already promoted.
	if s.UserSvc != nil {
		_ = s.UserSvc.PromoteToPartnerRole(ownerUserID)
	}

	return &partner, nil
}

func (s *Service) GetOwned(ownerUserID uint) ([]Partner, error) {
	return s.Repo.FindByOwnerID(ownerUserID)
}

func (s *Service) GetOwnedByID(ownerUserID uint, externalID string) (*Partner, error) {
	return s.Repo.FindByIDAndOwner(externalID, ownerUserID)
}

func (s *Service) UpdateOwned(ownerUserID uint, externalID string, req UpdatePartnerRequest) (*Partner, error) {
	partner, err := s.Repo.FindByIDAndOwner(externalID, ownerUserID)
	if err != nil {
		return nil, errors.New("listing not found")
	}

	if req.Name != nil {
		partner.Name = *req.Name
	}
	if req.Description != nil {
		partner.Description = *req.Description
	}
	if req.Category != nil {
		partner.Category = *req.Category
	}
	if req.Location != nil {
		partner.Location = *req.Location
	}
	if req.Address != nil {
		partner.Address = *req.Address
	}
	if req.Image != nil {
		partner.Image = *req.Image
	}
	if req.Phone != nil {
		partner.Phone = *req.Phone
	}
	if req.Website != nil {
		partner.Website = *req.Website
	}
	if req.Latitude != nil {
		partner.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		partner.Longitude = *req.Longitude
	}
	if req.Price != nil {
		partner.Price = *req.Price
	}

	// Approved listing reset to pending for re-review
	if partner.Status == StatusApproved {
		partner.Status = StatusPending
	}

	if err := s.Repo.Update(partner); err != nil {
		return nil, err
	}
	return partner, nil
}

func (s *Service) DeleteOwned(ownerUserID uint, externalID string) error {
	return s.Repo.DeleteByIDAndOwner(externalID, ownerUserID)
}
