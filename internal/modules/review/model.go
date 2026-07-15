package review

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

type Review struct {
	gorm.Model
	ExternalID    string  `gorm:"uniqueIndex;not null" json:"id"`
	UserID        string  `gorm:"index" json:"user_id"`
	DestinationID string  `gorm:"index" json:"destination_id"`
	UserName      string  `json:"user_name"`
	TravelerType  string  `gorm:"index" json:"traveler_type"` // Solo | Couple | Family | Friends
	Rating        int     `json:"rating"`
	Comment       string  `gorm:"type:text" json:"comment"`
	Images        JSONArr `gorm:"type:jsonb" json:"images"`
	Status        string  `gorm:"index" json:"status"`
}
