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
	VideoURL         string  `gorm:"column:video_url" json:"video_url"`

	// English translations (_en columns)
	NameEn             string  `gorm:"column:name_en" json:"name_en"`
	TaglineEn          string  `gorm:"column:tagline_en" json:"tagline_en"`
	DescriptionEn      string  `gorm:"type:text;column:description_en" json:"description_en"`
	StoryEn            string  `gorm:"type:text;column:story_en" json:"story_en"`
	BestTimeEn         string  `gorm:"column:best_time_en" json:"best_time_en"`
	FacilitiesEn       JSONArr `gorm:"type:jsonb;column:facilities_en" json:"facilities_en"`
	TravelTipsEn       JSONArr `gorm:"type:jsonb;column:travel_tips_en" json:"travel_tips_en"`
}

// Localize returns a copy of the destination with fields swapped based on locale.
// Default is Indonesian (main columns). English uses _en columns with fallback.
func (d *Destination) Localize(locale string) Destination {
	out := *d
	if locale == "en" {
		if d.NameEn != "" {
			out.Name = d.NameEn
		}
		if d.TaglineEn != "" {
			out.Tagline = d.TaglineEn
		}
		if d.DescriptionEn != "" {
			out.Description = d.DescriptionEn
		}
		if d.StoryEn != "" {
			out.Story = d.StoryEn
		}
		if d.BestTimeEn != "" {
			out.BestTime = d.BestTimeEn
		}
		if len(d.FacilitiesEn) > 0 {
			out.Facilities = d.FacilitiesEn
		}
		if len(d.TravelTipsEn) > 0 {
			out.TravelTips = d.TravelTipsEn
		}
	}
	return out
}

type UserDestination struct {
	gorm.Model
	UserID          uint   `gorm:"index;not null" json:"user_id"`
	DestinationSlug string `gorm:"index;not null" json:"destination_slug"`
	Status          string `gorm:"not null" json:"status"`
}

// DestinationResponse wraps Destination with computed badge fields.
// Badge is the single highest-priority badge (for card overlay display).
// Badges contains all badges that apply to this destination.
type DestinationResponse struct {
	Destination
	Badge  BadgeType   `json:"badge"`
	Badges []BadgeType `json:"badges"`
}

// ToResponse converts a Destination to DestinationResponse by computing its badges.
// trendingIDs is a set of external-IDs that the AI trending endpoint has elected
// as trending today.  Pass nil if the data is unavailable (no trending badge will be set).
func (d *Destination) ToResponse(trendingIDs map[string]bool) DestinationResponse {
	badges := ResolveBadges(*d, trendingIDs)
	primary := PrimaryBadge(*d, trendingIDs)
	if badges == nil {
		badges = []BadgeType{}
	}
	return DestinationResponse{
		Destination: *d,
		Badge:       primary,
		Badges:      badges,
	}
}
