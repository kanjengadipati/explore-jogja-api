package article

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

type Article struct {
	gorm.Model
	ExternalID          string     `gorm:"uniqueIndex;not null" json:"id"`
	Slug                string     `gorm:"uniqueIndex;not null" json:"slug"`
	Title               string     `gorm:"not null" json:"title"`
	TitleEn             string     `gorm:"column:title_en" json:"title_en"`
	Excerpt             string     `gorm:"type:text" json:"excerpt"`
	ExcerptEn           string     `gorm:"type:text;column:excerpt_en" json:"excerpt_en"`
	Content             string     `gorm:"type:text" json:"content"`
	ContentEn           string     `gorm:"type:text;column:content_en" json:"content_en"`
	CoverImage          string     `gorm:"column:cover_image" json:"cover_image"`
	Category            string     `gorm:"index" json:"category"`
	Tags                JSONArr    `gorm:"type:jsonb" json:"tags"`
	Author              string     `json:"author"`
	Status              string     `gorm:"index;default:draft" json:"status"`
	PublishedAt         *time.Time `gorm:"column:published_at" json:"published_at"`
	SeoTitle            string     `gorm:"column:seo_title" json:"seo_title"`
	SeoTitleEn          string     `gorm:"column:seo_title_en" json:"seo_title_en"`
	SeoDescription      string     `gorm:"type:text;column:seo_description" json:"seo_description"`
	SeoDescriptionEn    string     `gorm:"type:text;column:seo_description_en" json:"seo_description_en"`
	SeoKeywords         string     `gorm:"type:text;column:seo_keywords" json:"seo_keywords"`
	SeoKeywordsEn       string     `gorm:"type:text;column:seo_keywords_en" json:"seo_keywords_en"`
	OgImage             string     `gorm:"column:og_image" json:"og_image"`
	ReadTimeMinutes     int        `gorm:"column:read_time_minutes;default:5" json:"read_time_minutes"`
}

// Localize returns a copy of the article with fields swapped based on locale.
// Default is Indonesian. English uses _en columns with fallback to Indonesian.
func (a *Article) Localize(locale string) Article {
	out := *a
	if locale == "en" {
		if a.TitleEn != "" {
			out.Title = a.TitleEn
		}
		if a.ExcerptEn != "" {
			out.Excerpt = a.ExcerptEn
		}
		if a.ContentEn != "" {
			out.Content = a.ContentEn
		}
		if a.SeoTitleEn != "" {
			out.SeoTitle = a.SeoTitleEn
		}
		if a.SeoDescriptionEn != "" {
			out.SeoDescription = a.SeoDescriptionEn
		}
		if a.SeoKeywordsEn != "" {
			out.SeoKeywords = a.SeoKeywordsEn
		}
	}
	return out
}
