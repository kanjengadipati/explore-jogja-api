package business

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PartnerMirrorSyncer lets business writes mirror back onto the legacy partners
// row so both endpoint families stay consistent during the transition (Phase 3
// reverse dual-write). Implemented by partner.Service. Temporary scaffolding,
// retired in Phase 6.
type PartnerMirrorSyncer interface {
	SyncMirrorFromBusiness(m PartnerMirror) error
}

type Service struct {
	Repo        Repository
	PartnerSync PartnerMirrorSyncer
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

// PartnerMirror is a denormalized snapshot of a partners row used by the
// Phase 1 dual-write hook. Kept as primitives so this package never has to
// import the partner package.
type PartnerMirror struct {
	ExternalID              string
	Name                    string
	Description             string
	Category                string
	Phone                   string
	Website                 string
	Status                  string
	RejectionReason         string
	SubmittedAt             *time.Time
	ReviewedAt              *time.Time
	ReviewedBy              *uint
	OwnerUserID             *uint
	LegacyPartnerExternalID *string
}

// SyncFromPartner upserts the businesses row mirroring a partners row, joined
// on legacy_partner_external_id, and mirrors owner_user_id into business_owners.
// Temporary Phase 1 scaffolding; retired in Phase 6.
func (s *Service) SyncFromPartner(m PartnerMirror) error {
	existing, err := s.Repo.FindByLegacyPartner(m.ExternalID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		b := Business{
			ExternalID:              "biz_" + m.ExternalID,
			Name:                    m.Name,
			Description:             m.Description,
			Category:                m.Category,
			Phone:                   m.Phone,
			Website:                 m.Website,
			Status:                  m.Status,
			RejectionReason:         m.RejectionReason,
			SubmittedAt:             m.SubmittedAt,
			ReviewedAt:              m.ReviewedAt,
			ReviewedBy:              m.ReviewedBy,
			LegacyPartnerExternalID: &m.ExternalID,
		}
		if err := s.Repo.Create(&b); err != nil {
			return err
		}
		return s.mirrorOwner(b.ID, m.OwnerUserID)
	}
	if err != nil {
		return err
	}

	existing.Name = m.Name
	existing.Description = m.Description
	existing.Category = m.Category
	existing.Phone = m.Phone
	existing.Website = m.Website
	existing.Status = m.Status
	existing.RejectionReason = m.RejectionReason
	existing.SubmittedAt = m.SubmittedAt
	existing.ReviewedAt = m.ReviewedAt
	existing.ReviewedBy = m.ReviewedBy
	if err := s.Repo.Update(existing); err != nil {
		return err
	}
	return s.mirrorOwner(existing.ID, m.OwnerUserID)
}

func (s *Service) mirrorOwner(businessID uint, ownerUserID *uint) error {
	if ownerUserID == nil {
		return nil
	}
	return s.Repo.UpsertOwner(businessID, *ownerUserID)
}

// DeleteForPartner soft-deletes the businesses row mirrored from a partners row.
func (s *Service) DeleteForPartner(legacyExternalID string) error {
	return s.Repo.SoftDeleteByLegacyPartner(legacyExternalID)
}

// mirrorToPartner pushes business edits back onto the mirrored partners row.
// The partner write re-fires the Phase 1 forward hook, which re-applies the
// same values onto the business row — so the round-trip terminates after one
// hop and both rows stay identical.
func (s *Service) mirrorToPartner(b *Business) {
	if s.PartnerSync == nil || b.LegacyPartnerExternalID == nil {
		return
	}
	legacyID := *b.LegacyPartnerExternalID
	_ = s.PartnerSync.SyncMirrorFromBusiness(PartnerMirror{
		ExternalID:      legacyID,
		Name:            b.Name,
		Description:     b.Description,
		Category:        b.Category,
		Phone:           b.Phone,
		Website:         b.Website,
		Status:          b.Status,
		RejectionReason: b.RejectionReason,
		SubmittedAt:     b.SubmittedAt,
		ReviewedAt:      b.ReviewedAt,
		ReviewedBy:      b.ReviewedBy,
	})
}

// --- Self-service dashboard ---

type CreateBusinessRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	AvatarURL   string `json:"avatar_url"`
}

func (s *Service) CreateOwned(userID uint, req CreateBusinessRequest) (*Business, error) {
	extID := fmt.Sprintf("biz_%d", time.Now().UnixNano())
	now := time.Now()

	b := Business{
		ExternalID:  extID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		AvatarURL:   req.AvatarURL,
		Status:      StatusPending,
		SubmittedAt: &now,
	}

	if err := s.Repo.Create(&b); err != nil {
		return nil, err
	}

	if err := s.Repo.UpsertOwner(b.ID, userID); err != nil {
		return nil, err
	}

	s.mirrorToPartner(&b)
	return &b, nil
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
	s.mirrorToPartner(b)
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

// SetStatus applies an admin decision (approve/reject/suspend) and mirrors it
// back onto the legacy partners row so both endpoint families stay consistent.
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
	s.mirrorToPartner(b)
	return b, nil
}
