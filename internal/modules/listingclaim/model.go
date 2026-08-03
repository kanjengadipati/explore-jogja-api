package listingclaim

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// Valid listing_type values for listing_claims. Mirrors the DB CHECK constraint.
var ValidListingTypes = []string{
	"destination", "hotel", "restaurant", "souvenir", "rental", "guide", "event",
	"culinary", "shopping", "accommodation", "attraction",
}

// listingTable maps a listing_type to its table name. The approve flow uses this
// to set business_id on the right table; only known types are accepted.
var listingTable = map[string]string{
	"destination":   "destinations",
	"attraction":    "destinations",
	"hotel":         "hotels",
	"accommodation": "hotels",
	"restaurant":    "restaurants",
	"culinary":      "restaurants",
	"souvenir":      "souvenirs",
	"shopping":      "souvenirs",
	"rental":        "rentals",
	"guide":         "guides",
	"event":         "events",
}

type ListingClaim struct {
	gorm.Model
	ExternalID        string `gorm:"uniqueIndex;not null" json:"id"`
	BusinessID        uint   `gorm:"index;not null" json:"business_id"`
	ListingType       string `gorm:"size:50;not null" json:"listing_type"`
	ListingExternalID string `gorm:"not null" json:"listing_external_id"`

	Status          string `gorm:"size:20;not null;default:pending;index" json:"status"`
	RejectionReason string `gorm:"type:text" json:"rejection_reason"`

	SubmittedAt *time.Time `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	ReviewedBy  *uint      `json:"reviewed_by"`
}

type SearchResult struct {
	ListingType string `json:"listing_type"`
	ExternalID  string `json:"id"`
	Name        string `json:"name"`
}
