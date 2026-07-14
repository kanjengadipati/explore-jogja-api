package partner

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

type Partner struct {
	gorm.Model
	ExternalID string  `gorm:"uniqueIndex;not null" json:"id"`
	Name       string  `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	Category   string  `gorm:"index" json:"category"`
	Location   string  `json:"location"`
	Address    string  `json:"address"`
	Image      string  `json:"image"`
	Rating     float64 `json:"rating"`
	Price      string  `json:"price"`
	Distance   string  `json:"distance"`
	Phone      string  `json:"phone"`
	Website    string  `json:"website"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
}
