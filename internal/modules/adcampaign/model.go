package adcampaign

import (
	"time"

	"gorm.io/gorm"
)

const (
	PlacementHomepageHeroAICard     = "homepage_hero_aicard"
	PlacementHomepageHeroTrending   = "homepage_hero_trending"
	PlacementHomepageCategoryBanner = "homepage_category_banner"
	PlacementListingTop             = "listing_top"
	PlacementListingNative          = "listing_native"
	PlacementDestinationDetail      = "destination_detail"
)

type AdCampaign struct {
	gorm.Model
	ExternalID         string  `gorm:"uniqueIndex;not null" json:"id"`
	BusinessExternalID *string `gorm:"index" json:"business_external_id"`
	// PartnerName is the legacy free-text column, still NOT NULL in the DB
	// until Phase 4 step 4 drops it. Kept on the model so the admin create flow
	// can keep satisfying the constraint while filling business_external_id.
	PartnerName string `gorm:"not null" json:"partner_name,omitempty"`
	// BusinessName is read-only, populated via a LEFT JOIN in the repository
	// (never a DB column).
	BusinessName string `gorm:"-" json:"business_name,omitempty"`
	Placement    string `gorm:"index;not null" json:"placement"`
	ImageURL     string `gorm:"not null" json:"image_url"`
	TargetURL    string `gorm:"not null" json:"target_url"`
	Category     string `gorm:"index" json:"category"`

	StartAt time.Time `json:"start_at"`
	EndAt   time.Time `json:"end_at"`

	Weight int `gorm:"default:1" json:"weight"`

	Impressions int64 `json:"impressions"`
	Clicks      int64 `json:"clicks"`

	IsActive bool `gorm:"default:true;index" json:"is_active"`

	PriceAmount   float64 `json:"price_amount"`
	PriceCurrency string  `gorm:"default:IDR" json:"price_currency"`
	PaymentStatus string  `gorm:"default:pending" json:"payment_status"`
}

type HouseAd struct {
	gorm.Model
	ExternalID  string `gorm:"uniqueIndex;not null" json:"id"`
	Placement   string `gorm:"uniqueIndex;not null" json:"placement"`
	Headline    string `gorm:"not null" json:"headline"`
	HeadlineEn  string `gorm:"column:headline_en" json:"headline_en"`
	Subline     string `gorm:"type:text" json:"subline"`
	SublineEn   string `gorm:"type:text;column:subline_en" json:"subline_en"`
	CTALabel    string `gorm:"not null" json:"cta_label"`
	CTALabelEn  string `gorm:"column:cta_label_en" json:"cta_label_en"`
	ImageURL    string `json:"image_url"`
	TargetURL   string `gorm:"not null" json:"target_url"`
	IsEnabled   bool   `gorm:"index" json:"is_enabled"`
}

func (a *AdCampaign) IsLive(now time.Time) bool {
	if !a.IsActive {
		return false
	}
	if a.PaymentStatus != "paid" {
		return false
	}
	if !a.StartAt.IsZero() && now.Before(a.StartAt) {
		return false
	}
	if !a.EndAt.IsZero() && now.After(a.EndAt) {
		return false
	}
	return true
}
