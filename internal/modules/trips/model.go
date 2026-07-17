package trips

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// JSONDays stores the per-day itinerary as a JSONB column.
type JSONDays []TripDayJSON

type TripDayJSON struct {
	DayNumber    int      `json:"dayNumber"`
	DestIDs      []string `json:"destinationIds"`
	Notes        string   `json:"notes"`
}

func (j JSONDays) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONDays) Scan(value interface{}) error {
	if value == nil {
		*j = JSONDays{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("JSONDays: unsupported type %T", value)
	}
	return json.Unmarshal(b, j)
}

// Trip is the GORM model for the trips table.
type Trip struct {
	gorm.Model
	ExternalID   string   `gorm:"uniqueIndex;not null"                json:"id"`
	UserID       uint     `gorm:"index;not null"                      json:"user_id"`
	Title        string   `gorm:"not null;default:'My Trip'"          json:"title"`
	StartDate    string   `gorm:"type:date"                           json:"start_date"`
	EndDate      string   `gorm:"type:date"                           json:"end_date"`
	DurationDays int      `gorm:"not null;default:1"                  json:"duration_days"`
	Days         JSONDays `gorm:"type:jsonb;not null;default:'[]'"    json:"days"`
	Notes        string   `gorm:"type:text"                           json:"notes"`
	Status       string   `gorm:"index;not null;default:'draft'"      json:"status"`
}
