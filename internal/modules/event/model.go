package event

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

type Event struct {
	gorm.Model
	ExternalID    string  `gorm:"uniqueIndex;not null" json:"id"`
	Title         string  `gorm:"not null" json:"title"`
	Description   string  `gorm:"type:text" json:"description"`
	Location      string  `json:"location"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
	ImageURL      string  `json:"image_url"`
	Category      string  `gorm:"index" json:"category"`
	Status        string  `gorm:"index" json:"status"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	MaxAttendees  int     `json:"max_attendees"`
	TicketPrice   string  `json:"ticket_price"`
	Organizer     string  `json:"organizer"`
	Highlights    JSONArr `gorm:"type:jsonb" json:"highlights"`
	DestinationID string  `json:"destination_id"`
}
