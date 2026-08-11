package event

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
