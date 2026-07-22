package staging

import (
	"time"

	"gorm.io/gorm"
)

type StagingDestination struct {
	gorm.Model
	Source         string `json:"source"`
	ProviderID     string `json:"provider_id"`
	Name           string `json:"name"`
	Description    string `gorm:"type:text" json:"description"`
	Latitude       string `json:"latitude"`
	Longitude      string `json:"longitude"`
	Address        string `json:"address"`
	Category       string `json:"category"`
	Images         string `gorm:"type:text" json:"images"` // JSON string
	RawData        string `gorm:"type:text" json:"raw_data"`
	Status         string `gorm:"default:'pending'" json:"status"` // pending, approved, rejected
}

type StagingEvent struct {
	gorm.Model
	Source         string `json:"source"`
	ProviderID     string `json:"provider_id"`
	Title          string `json:"title"`
	Description    string `gorm:"type:text" json:"description"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	Location       string `json:"location"`
	RawData        string `gorm:"type:text" json:"raw_data"`
	Status         string `gorm:"default:'pending'" json:"status"`
}
