package restaurant

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type JSONArr []interface{}

func (j JSONArr) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONArr) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArr, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONArr: value is not []byte")
	}
	return json.Unmarshal(bytes, j)
}

type Restaurant struct {
	gorm.Model
	ExternalID   string  `gorm:"uniqueIndex;not null" json:"id"`
	Name         string  `gorm:"not null" json:"name"`
	Description  string  `gorm:"type:text" json:"description"`
	Location     string  `json:"location"`
	Address      string  `json:"address"`
	CuisineType  string  `json:"cuisine_type"`
	PriceRange   string  `json:"price_range"`
	Images       JSONArr `gorm:"type:jsonb" json:"images"`
	OpeningHours string  `json:"opening_hours"`
	Phone        string  `json:"phone"`
	Rating       float64 `json:"rating"`
	ReviewCount  int     `json:"review_count"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
}
