package business

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusDraft     = "draft"
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusSuspended = "suspended"

	RoleOwner = "owner"
	RoleAdmin = "admin"
)

// Business is the identity of a business account (the Partner -> Business
// migration's first split). It deliberately carries no sponsorship or listing
// ownership: entitlements live in the subscriptions table and ownership lives
// in business_owners + listing claims.
type Business struct {
	gorm.Model
	ID                      uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ExternalID              string     `gorm:"uniqueIndex;not null" json:"external_id"`
	Name                    string     `gorm:"not null" json:"name"`
	Description             string     `gorm:"type:text" json:"description"`
	Category                string     `gorm:"not null;index" json:"category"`
	Phone                   string     `json:"phone"`
	Address                 string     `json:"address"`
	Email                   string     `json:"email"`
	Website                 string     `json:"website"`
	AvatarURL               string     `json:"avatar_url"`
	Status                  string     `gorm:"size:20;not null;default:pending;index" json:"status"`
	RejectionReason         string     `gorm:"type:text" json:"rejection_reason"`
	SubmittedAt             *time.Time `json:"submitted_at"`
	ReviewedAt              *time.Time `json:"reviewed_at"`
	ReviewedBy              *uint      `json:"reviewed_by"`
	LegacyPartnerExternalID *string    `gorm:"index" json:"legacy_partner_external_id"`

	Owners       []BusinessOwner       `gorm:"foreignKey:BusinessID" json:"owners,omitempty"`
	ServiceAreas []BusinessServiceArea `gorm:"foreignKey:BusinessID" json:"service_areas,omitempty"`
}

// BusinessServiceArea is one region (kabupaten/kota DIY) a business operates in.
// A business can have multiple rows — e.g. a tour guide covering both Sleman and
// Bantul. Region values are validated at the application layer (see
// ValidServiceAreaRegions in service.go), not via a DB CHECK, so the list can be
// extended without a schema migration if coverage expands beyond DIY later.
type BusinessServiceArea struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	BusinessID uint   `gorm:"uniqueIndex:idx_business_service_area_unique;not null" json:"business_id"`
	Region     string `gorm:"uniqueIndex:idx_business_service_area_unique;not null" json:"region"`
}

// BusinessOwner links a user account to a business. The DB-side unique index
// guarantees one row per (business, user). This table mirrors partners.owner_user_id.
type BusinessOwner struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	BusinessID uint      `gorm:"uniqueIndex:idx_business_owners_unique;not null" json:"business_id"`
	UserID     uint      `gorm:"uniqueIndex:idx_business_owners_unique;index;not null" json:"user_id"`
	Role       string    `gorm:"size:20;not null;default:owner" json:"role"`
	InvitedBy  *uint     `json:"invited_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// OwnedListing is a lightweight projection of a listing claimed by a business,
// aggregated across the 7 claimable listing tables. Hotels/guides/restaurants/
// souvenirs/rentals have no status column, so Status is empty for those.
type OwnedListing struct {
	ListingType string `json:"listing_type"`
	ExternalID  string `json:"id"`
	Name        string `json:"name"`
	Status      string `json:"status,omitempty"`
}

// PublicBusinessResponse is the safe, public-facing projection of a Business
// record — it omits internal fields (rejection_reason, reviewed_by, legacy IDs)
// and uses ExternalID as the public ID.
type PublicBusinessResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Phone       string `json:"phone"`
	Website     string `json:"website"`
	AvatarURL   string `json:"avatar_url"`
	Status      string `json:"status"`
}

func ToPublicBusinessResponse(b Business) PublicBusinessResponse {
	return PublicBusinessResponse{
		ID:          b.ExternalID,
		Name:        b.Name,
		Description: b.Description,
		Category:    b.Category,
		Phone:       b.Phone,
		Website:     b.Website,
		AvatarURL:   b.AvatarURL,
		Status:      b.Status,
	}
}

func ToPublicBusinessResponses(businesses []Business) []PublicBusinessResponse {
	out := make([]PublicBusinessResponse, 0, len(businesses))
	for _, b := range businesses {
		out = append(out, ToPublicBusinessResponse(b))
	}
	return out
}
