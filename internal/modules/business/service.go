package business

import (
	"errors"
	"fmt"
	"time"
)

type Service struct {
	Repo    Repository
	UserSvc PartnerPromoter
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

// PartnerPromoter upgrades a user to the partner role once they own a
// business. Implemented by user.Service.
type PartnerPromoter interface {
	PromoteToPartnerRole(userID uint) error
	// GetIDByReferralCode and SetReferredBySales support the optional sales
	// referral flow — a business can still be created fine if the code is
	// empty or unknown (see CreateOwned).
	GetIDByReferralCode(code string) (uint, error)
	SetReferredBySales(userID, salesUserID uint) error
}

// --- Self-service dashboard ---

// ValidServiceAreaRegions is the fixed whitelist for CreateBusinessRequest.Regions.
// Kept as a plain slice (not a Gin binding "oneof" tag) because region names
// contain spaces ("Kota Yogyakarta"), which oneof handles awkwardly.
var ValidServiceAreaRegions = []string{
	"Kota Yogyakarta", "Sleman", "Bantul", "Kulon Progo", "Gunungkidul", "Near Yogyakarta",
}

func IsValidServiceAreaRegion(region string) bool {
	for _, r := range ValidServiceAreaRegions {
		if r == region {
			return true
		}
	}
	return false
}

// CreateBusinessRequest — Name, Category, Phone, Address, and at least one
// Region are required. Phone is the primary channel admin uses to verify
// ownership before a listing claim is approved. Address is the specific street
// address (verification use, not shown for public filtering). Regions is
// separate from Address — it's the region-level service area (kabupaten/kota)
// used for search/filter, and a business can serve more than one.
// Email and Description stay optional; they can be filled in later via
// profile edit after approval.
type CreateBusinessRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category" binding:"required"`
	Phone       string   `json:"phone" binding:"required"`
	Address     string   `json:"address" binding:"required"`
	Regions     []string `json:"regions" binding:"required,min=1"`
	Email       string   `json:"email"`
	Website     string   `json:"website"`
	AvatarURL   string   `json:"avatar_url"`
	// ReferralCode is optional — set when the partner signed up through a
	// sales referral link/code. An unknown or invalid code is ignored on
	// purpose so a typo never blocks registration.
	ReferralCode string `json:"referral_code,omitempty"`
}

func (s *Service) CreateOwned(userID uint, req CreateBusinessRequest) (*Business, error) {
	extID := fmt.Sprintf("biz_%d", time.Now().UnixNano())
	now := time.Now()

	for _, region := range req.Regions {
		if !IsValidServiceAreaRegion(region) {
			return nil, fmt.Errorf("wilayah tidak dikenal: %s", region)
		}
	}

	b := Business{
		ExternalID:  extID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Phone:       req.Phone,
		Address:     req.Address,
		Email:       req.Email,
		Website:     req.Website,
		AvatarURL:   req.AvatarURL,
		Status:      StatusApproved,
		SubmittedAt: &now,
		ReviewedAt:  &now,
	}

	if err := s.Repo.CreateWithServiceAreas(&b, req.Regions); err != nil {
		return nil, err
	}

	if err := s.Repo.UpsertOwner(b.ID, userID, RoleOwner); err != nil {
		return nil, err
	}

	if s.UserSvc != nil {
		if err := s.UserSvc.PromoteToPartnerRole(userID); err != nil {
			return nil, err
		}
		if req.ReferralCode != "" {
			if salesID, err := s.UserSvc.GetIDByReferralCode(req.ReferralCode); err == nil {
				// Best-effort — an unknown/invalid code should never block
				// business creation, so any error here is swallowed.
				_ = s.UserSvc.SetReferredBySales(userID, salesID)
			}
		}
	}

	return &b, nil
}

// FindSimilarApprovedName returns approved businesses whose names are similar
// to the given query (case-insensitive ILIKE). Used for the global name-dedup
// check at registration time. Returns an empty slice (not an error) when the
// query is too short, to avoid blocking the submit step.
func (s *Service) FindSimilarApprovedName(query string) ([]Business, error) {
	if len(query) < 3 {
		return []Business{}, nil
	}
	return s.Repo.FindSimilarByName(query, StatusApproved)
}

func (s *Service) GetOwned(userID uint) ([]Business, error) {
	return s.Repo.FindByOwner(userID)
}

func (s *Service) GetOwnedByID(userID uint, externalID string) (*Business, error) {
	return s.Repo.FindByIDAndOwner(externalID, userID)
}

type UpdateBusinessRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Website     *string `json:"website"`
	AvatarURL   *string `json:"avatar_url"`
}

func (s *Service) UpdateOwned(userID uint, externalID string, req UpdateBusinessRequest) (*Business, error) {
	b, err := s.Repo.FindByIDAndOwner(externalID, userID)
	if err != nil {
		return nil, errors.New("business not found")
	}

	if req.Name != nil {
		b.Name = *req.Name
	}
	if req.Description != nil {
		b.Description = *req.Description
	}
	if req.Category != nil {
		b.Category = *req.Category
	}
	if req.Phone != nil {
		b.Phone = *req.Phone
	}
	if req.Email != nil {
		b.Email = *req.Email
	}
	if req.Website != nil {
		b.Website = *req.Website
	}
	if req.AvatarURL != nil {
		b.AvatarURL = *req.AvatarURL
	}

	// Approved business reset to pending for re-review (mirrors partner flow)
	if b.Status == StatusApproved {
		b.Status = StatusPending
	}

	if err := s.Repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) GetListings(businessID uint) ([]OwnedListing, error) {
	return s.Repo.GetListings(businessID)
}

// --- Admin ---

func (s *Service) GetAllAny() ([]Business, error) {
	return s.Repo.FindAll()
}

func (s *Service) GetByIDAny(externalID string) (*Business, error) {
	return s.Repo.FindByID(externalID)
}

func (s *Service) GetByStatus(status string) ([]Business, error) {
	return s.Repo.FindByStatus(status)
}

// SetStatus applies an admin decision (approve/reject/suspend).
func (s *Service) SetStatus(externalID, status, reason string, adminUserID uint) (*Business, error) {
	b, err := s.Repo.FindByID(externalID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	b.Status = status
	b.RejectionReason = reason
	b.ReviewedAt = &now
	b.ReviewedBy = &adminUserID

	if err := s.Repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}
