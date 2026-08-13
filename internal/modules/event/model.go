package event

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	ExternalID       string    `gorm:"uniqueIndex;not null" json:"id"`
	Title            string    `gorm:"not null" json:"title"`
	Description      string    `gorm:"type:text" json:"description"`
	Location         string    `json:"location"`
	StartDate        string    `json:"start_date"`
	EndDate          string    `json:"end_date"`
	ImageURL         string    `json:"image_url"`
	Images           JSONArr   `gorm:"type:jsonb" json:"images"`
	Category         string    `gorm:"index" json:"category"`
	Status           string    `gorm:"index" json:"status"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	MaxAttendees     int       `json:"max_attendees"`
	TicketPrice      string    `json:"ticket_price"`
	Organizer        string    `json:"organizer"`
	Highlights       JSONArr   `gorm:"type:jsonb" json:"highlights"`
	DestinationID    string    `json:"destination_id"`
	VideoURL         string    `gorm:"column:video_url" json:"video_url"`
	TitleEn          string    `json:"title_en"`
	DescriptionEn    string    `gorm:"type:text" json:"description_en"`
	SeoTitle         string    `json:"seo_title"`
	SeoTitleEn       string    `json:"seo_title_en"`
	SeoDescription   string    `gorm:"type:text" json:"seo_description"`
	SeoDescriptionEn string    `gorm:"type:text" json:"seo_description_en"`
	SeoKeywords      string    `gorm:"type:text" json:"seo_keywords"`
	SeoKeywordsEn    string    `gorm:"type:text" json:"seo_keywords_en"`
	OgImageUrl       string    `json:"og_image_url"`
	LastScrapedAt    time.Time `gorm:"column:last_scraped_at" json:"-"`
	// ContentScore is the deterministic rubric score (0–100) calculated from field
	// completeness. Mirrors destination.ContentScore (see internal/modules/destination/quality.go).
	ContentScore   int    `gorm:"default:0" json:"content_score"`
	ContentVerdict string `gorm:"default:''" json:"content_verdict"` // EXCELLENT | GOOD | NEEDS WORK
}

// UnmarshalJSON accepts latitude/longitude as either a JSON number or a
// JSON string. The admin dashboard submits coordinates as strings (e.g.
// "-7.7928"), which encoding/json normally rejects for a float64 field —
// that surfaced as a 500 on create. Both forms are coerced to float64.
func (e *Event) UnmarshalJSON(data []byte) error {
	type EventAlias Event
	var aux struct {
		EventAlias
		Latitude  json.RawMessage `json:"latitude"`
		Longitude json.RawMessage `json:"longitude"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*e = Event(aux.EventAlias)

	var err error
	if e.Latitude, err = parseFlexibleFloat(aux.Latitude, "latitude"); err != nil {
		return err
	}
	if e.Longitude, err = parseFlexibleFloat(aux.Longitude, "longitude"); err != nil {
		return err
	}
	return nil
}

func parseFlexibleFloat(raw json.RawMessage, field string) (float64, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return 0, nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return 0, fmt.Errorf("%s: %w", field, err)
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return 0, nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", field, err)
		}
		return v, nil
	}
	var v float64
	if err := json.Unmarshal(trimmed, &v); err != nil {
		return 0, fmt.Errorf("%s: %w", field, err)
	}
	return v, nil
}
