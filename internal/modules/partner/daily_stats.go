package partner

import (
	"time"
)

type PartnerDailyStats struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	PartnerExternalID string    `gorm:"not null;index" json:"partner_external_id"`
	Date              time.Time `gorm:"type:date;not null;index" json:"date"`
	Impressions       int       `gorm:"default:0" json:"impressions"`
	Clicks            int       `gorm:"default:0" json:"clicks"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
