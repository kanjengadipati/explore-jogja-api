package destination

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONMap: value is not []byte")
	}
	return json.Unmarshal(bytes, j)
}

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

type WeatherInfo struct {
	Temp      string `json:"temp"`
	Condition string `json:"condition"`
	Status    string `json:"status"`
}

type Destination struct {
	gorm.Model
	ExternalID       string  `gorm:"uniqueIndex;not null" json:"id"`
	Name             string  `gorm:"not null" json:"name"`
	Tagline          string  `json:"tagline"`
	Category         string  `gorm:"index" json:"category"`
	Location         string  `json:"location"`
	SubRegion        string  `gorm:"index" json:"sub_region"`
	Images           JSONArr `gorm:"type:jsonb" json:"images"`
	Rating           float64 `json:"rating"`
	ReviewCount      int     `json:"review_count"`
	Description      string  `gorm:"type:text" json:"description"`
	Story            string  `gorm:"type:text" json:"story"`
	TicketPrice      string  `json:"ticket_price"`
	OpeningHours     string  `json:"opening_hours"`
	Facilities       JSONArr `gorm:"type:jsonb" json:"facilities"`
	TravelTips       JSONArr `gorm:"type:jsonb" json:"travel_tips"`
	BestTime         string  `json:"best_time"`
	Weather          JSONMap `gorm:"type:jsonb" json:"weather"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Reviews          JSONArr `gorm:"type:jsonb" json:"reviews"`
	Partners         JSONArr `gorm:"type:jsonb" json:"partners"`
	FAQs             JSONArr `gorm:"type:jsonb;column:faqs" json:"faqs"`
	GoogleMapsURL    string  `json:"google_maps_url"`
	GoogleReviewCount int    `json:"google_review_count"`
	SeoTitle         string  `json:"seo_title"`
	SeoKeywords      string  `gorm:"type:text" json:"seo_keywords"`
	SeoDescription   string  `gorm:"type:text" json:"seo_description"`
	OgImageUrl       string  `json:"og_image_url"`
}
