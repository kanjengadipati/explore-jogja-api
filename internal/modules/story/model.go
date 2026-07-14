package story

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

type Story struct {
	gorm.Model
	ExternalID    string `gorm:"uniqueIndex;not null" json:"id"`
	UserID        string `gorm:"index" json:"user_id"`
	Title         string `gorm:"not null" json:"title"`
	Content       string `gorm:"type:text" json:"content"`
	Images        JSONArr `gorm:"type:jsonb" json:"images"`
	DestinationIDs JSONArr `gorm:"type:jsonb" json:"destination_ids"`
	Likes         int     `json:"likes"`
	Status        string  `gorm:"index" json:"status"`
}
