package hotel

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

type Hotel struct {
	gorm.Model
	ExternalID    string  `gorm:"uniqueIndex;not null" json:"id"`
	Name          string  `gorm:"not null" json:"name"`
	Description   string  `gorm:"type:text" json:"description"`
	Location      string  `json:"location"`
	Address       string  `json:"address"`
	Stars         int     `json:"stars"`
	PricePerNight string  `json:"price_per_night"`
	Images        JSONArr `gorm:"type:jsonb" json:"images"`
	Amenities     JSONArr `gorm:"type:jsonb" json:"amenities"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	Website       string  `json:"website"`
	Rating        float64 `json:"rating"`
	ReviewCount   int     `json:"review_count"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
}
