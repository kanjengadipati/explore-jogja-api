package listingclaim

import (
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	"pleco-api/internal/modules/business"
)

var ErrNotBusinessOwner = errors.New("user is not an owner of this business")

type Service struct {
	Repo    Repository
	BizRepo business.Repository
}

func NewService(repo Repository, bizRepo business.Repository) *Service {
	return &Service{Repo: repo, BizRepo: bizRepo}
}

type SubmitRequest struct {
	BusinessExternalID string `json:"business_external_id" binding:"required"`
	ListingType        string `json:"listing_type" binding:"required"`
	ListingExternalID  string `json:"listing_external_id" binding:"required"`
	Role               string `json:"role"`
	EvidenceURL        string `json:"evidence_url"`
}

func (s *Service) Submit(req SubmitRequest, ownerUserID uint) (*ListingClaim, error) {
	if !slices.Contains(ValidListingTypes, req.ListingType) {
		return nil, ErrListingNotOwnable
	}

	// Resolve the external_id to the internal business record
	biz, err := s.BizRepo.FindByID(req.BusinessExternalID)
	if err != nil {
		return nil, ErrNotBusinessOwner
	}

	// Verify the caller is an owner of this business
	if s.BizRepo != nil {
		ok, err := s.BizRepo.IsOwner(biz.ID, ownerUserID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrNotBusinessOwner
		}
	}

	// Duplicate-submission guard: block if an active (pending/approved) claim
	// already exists for this exact listing — prevents spam submissions.
	existing, err := s.Repo.FindActiveByListing(req.ListingType, req.ListingExternalID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrClaimAlreadyPending
	}

	now := time.Now()
	claim := ListingClaim{
		ExternalID:        uuid.New().String(),
		BusinessID:        biz.ID,
		ListingType:       req.ListingType,
		ListingExternalID: req.ListingExternalID,
		Status:            StatusPending,
		Role:              req.Role,
		EvidenceURL:       req.EvidenceURL,
		SubmittedAt:       &now,
	}
	if err := s.Repo.Create(&claim); err != nil {
		return nil, err
	}
	return &claim, nil
}

func (s *Service) GetPending() ([]ListingClaim, error) {
	return s.Repo.FindPending()
}

func (s *Service) GetOwned(businessIDs []uint) ([]ListingClaim, error) {
	var all []ListingClaim
	for _, id := range businessIDs {
		claims, err := s.Repo.FindByBusiness(id)
		if err != nil {
			return nil, err
		}
		all = append(all, claims...)
	}
	return all, nil
}

func (s *Service) GetByID(externalID string) (*ListingClaim, error) {
	return s.Repo.FindByExternalID(externalID)
}

func (s *Service) Approve(externalID string, adminUserID uint) error {
	return s.Repo.Approve(externalID, adminUserID)
}

func (s *Service) Reject(externalID string, reason string, adminUserID uint) error {
	return s.Repo.Reject(externalID, reason, adminUserID)
}

func (s *Service) SearchListings(query string) ([]SearchResult, error) {
	if len(query) < 3 {
		return []SearchResult{}, nil
	}
	return s.Repo.SearchListings(query)
}
