package adcampaign

import (
	"time"

	"gorm.io/gorm"
)

const (
	PlacementHomepageHero      = "homepage_hero"
	PlacementListingTop        = "listing_top"
	PlacementListingNative     = "listing_native"
	PlacementDestinationDetail = "destination_detail"
)

type AdCampaign struct {
	gorm.Model
	ExternalID  string `gorm:"uniqueIndex;not null" json:"id"`
	PartnerName string `gorm:"not null" json:"partner_name"`
	Placement   string `gorm:"index;not null" json:"placement"`
	ImageURL    string `gorm:"not null" json:"image_url"`
	TargetURL   string `gorm:"not null" json:"target_url"`
	Category    string `gorm:"index" json:"category"`

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
	ExternalID string `gorm:"uniqueIndex;not null" json:"id"`
	Placement  string `gorm:"uniqueIndex;not null" json:"placement"`
	Headline   string `gorm:"not null" json:"headline"`
	Subline    string `gorm:"type:text" json:"subline"`
	CTALabel   string `gorm:"not null" json:"cta_label"`
	ImageURL   string `json:"image_url"`
	TargetURL  string `gorm:"not null" json:"target_url"`
	IsEnabled  bool   `gorm:"index" json:"is_enabled"`
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
