package promotion

import (
	"pleco-api/internal/utils"

	"gorm.io/gorm"
)

type JSONArr = utils.JSONArr

type Promotion struct {
	gorm.Model
	ExternalID  string  `gorm:"uniqueIndex;not null" json:"id"`
	Title       string  `gorm:"not null" json:"title"`
	Description string  `gorm:"type:text" json:"description"`
	Discount    string  `json:"discount"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	ImageURL    string  `json:"image_url"`
	Category    string  `gorm:"index" json:"category"`
	Status                  string  `gorm:"index" json:"status"`
	Code                    string  `json:"code"`
	LegacyPartnerExternalID *string `gorm:"index" json:"legacy_partner_external_id"`
	BusinessExternalID      *string `gorm:"index" json:"business_external_id"`
}
