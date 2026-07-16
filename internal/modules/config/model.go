package config

import "gorm.io/gorm"

type SiteConfig struct {
	gorm.Model
	Key      string `gorm:"uniqueIndex;not null" json:"key"`
	Value    string `gorm:"type:text" json:"value"`
	Category string `gorm:"index" json:"category"`
}
