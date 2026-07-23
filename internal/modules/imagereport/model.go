package imagereport

import "gorm.io/gorm"

type ImageReport struct {
	gorm.Model
	DestinationID   string `gorm:"index;not null" json:"destination_id"`
	DestinationName string `gorm:"-" json:"destination_name"`
	DestinationLoc  string `gorm:"-" json:"destination_location"`
	ImageURL        string `gorm:"type:text" json:"image_url"`
	UserID          uint   `gorm:"index" json:"user_id"`
	UserName        string `json:"user_name"`
	Reason          string `gorm:"not null" json:"reason"`
	Details         string `gorm:"type:text" json:"details"`
	Status          string `gorm:"default:'pending';index" json:"status"` // pending, resolved, dismissed
}
