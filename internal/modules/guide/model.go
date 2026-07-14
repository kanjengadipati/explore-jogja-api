package guide

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

type Guide struct {
	gorm.Model
	ExternalID     string  `gorm:"uniqueIndex;not null" json:"id"`
	Name           string  `gorm:"not null" json:"name"`
	Bio            string  `gorm:"type:text" json:"bio"`
	Specialization string  `json:"specialization"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	Rating         float64 `json:"rating"`
	ReviewCount    int     `json:"review_count"`
	Languages      JSONArr `gorm:"type:jsonb" json:"languages"`
	PricePerDay    string  `json:"price_per_day"`
	Avatar         string  `json:"avatar"`
}
