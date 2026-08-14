package event

import (
	"context"
	"time"

	"pleco-api/internal/cache"
)

type EventResponse struct {
	ID               uint      `json:"id_numeric"`
	ExternalID       string    `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Location         string    `json:"location"`
	StartDate        string    `json:"start_date"`
	EndDate          string    `json:"end_date"`
	ImageURL         string    `json:"image_url"`
	Images           JSONArr   `json:"images"`
	Category         string    `json:"category"`
	Status           string    `json:"status"`
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	MaxAttendees     int       `json:"max_attendees"`
	TicketPrice      string    `json:"ticket_price"`
	Organizer        string    `json:"organizer"`
	DestinationID    string    `json:"destination_id"`
	Highlights       JSONArr   `json:"highlights"`
	VideoURL         string    `json:"video_url"`
	Badge            string    `json:"badge"`
	Badges           []string  `json:"badges"`
	TitleEn          string    `json:"title_en"`
	DescriptionEn    string    `json:"description_en"`
	SeoTitle         string    `json:"seo_title"`
	SeoTitleEn       string    `json:"seo_title_en"`
	SeoDescription   string    `json:"seo_description"`
	SeoDescriptionEn string    `json:"seo_description_en"`
	SeoKeywords      string    `json:"seo_keywords"`
	SeoKeywordsEn    string    `json:"seo_keywords_en"`
	OgImageUrl       string    `json:"og_image_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (e *Event) ToResponse(locale string, trendingIDs map[string]bool) EventResponse {
	var badges []string

	// Localize title/description for English visitors, falling back to the
	// Indonesian original when no translation exists.
	title := e.Title
	description := e.Description
	if locale == "en" {
		if e.TitleEn != "" {
			title = e.TitleEn
		}
		if e.DescriptionEn != "" {
			description = e.DescriptionEn
		}
	}

	// 1. Cek Trending dari AI
	if trendingIDs[e.ExternalID] {
		badges = append(badges, "trending")
	}

	// 2. Kriteria dari Status & Kapasitas
	status := e.Status

	if status == "popular" || e.MaxAttendees > 500 {
		badges = append(badges, "populer")
	}
	if status == "limited" {
		badges = append(badges, "terbatas")
	}
	if status == "upcoming" {
		badges = append(badges, "akan_datang")
	}

	// Default fallback ke kategori
	if len(badges) == 0 && e.Category != "" {
		badges = append(badges, e.Category)
	}

	// Tentukan primary badge
	primaryBadge := ""
	if len(badges) > 0 {
		primaryBadge = badges[0]
	}

	return EventResponse{
		ID:               e.ID,
		ExternalID:       e.ExternalID,
		Title:            title,
		Description:      description,
		Location:         e.Location,
		StartDate:        e.StartDate,
		EndDate:          e.EndDate,
		ImageURL:         e.ImageURL,
		Images:           e.Images,
		Category:         e.Category,
		Status:           e.Status,
		Latitude:         e.Latitude,
		Longitude:        e.Longitude,
		MaxAttendees:     e.MaxAttendees,
		TicketPrice:      e.TicketPrice,
		Organizer:        e.Organizer,
		DestinationID:    e.DestinationID,
		Highlights:       e.Highlights,
		VideoURL:         e.VideoURL,
		Badge:            primaryBadge,
		Badges:           badges,
		TitleEn:          e.TitleEn,
		DescriptionEn:    e.DescriptionEn,
		SeoTitle:         e.SeoTitle,
		SeoTitleEn:       e.SeoTitleEn,
		SeoDescription:   e.SeoDescription,
		SeoDescriptionEn: e.SeoDescriptionEn,
		SeoKeywords:      e.SeoKeywords,
		SeoKeywordsEn:    e.SeoKeywordsEn,
		OgImageUrl:       e.OgImageUrl,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

// loadTrendingIDs reads the AI-selected trending event IDs from Redis.
func loadTrendingIDs(cacheStore cache.Store, locale string) map[string]bool {
	var ids []string
	ok, err := cacheStore.GetJSON(context.Background(), cache.KeyAITrendingEventIDs(locale), &ids)
	if err != nil || !ok || len(ids) == 0 {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
