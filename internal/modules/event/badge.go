package event

import (
	"context"
	"pleco-api/internal/cache"
)

type EventResponse struct {
	ID            uint        `json:"id_numeric"`
	ExternalID    string      `json:"id"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Location      string      `json:"location"`
	StartDate     string      `json:"start_date"`
	EndDate       string      `json:"end_date"`
	ImageURL      string      `json:"image_url"`
	Category      string      `json:"category"`
	Status        string      `json:"status"`
	Latitude      float64     `json:"latitude"`
	Longitude     float64     `json:"longitude"`
	MaxAttendees  int         `json:"max_attendees"`
	TicketPrice   string      `json:"ticket_price"`
	Organizer     string      `json:"organizer"`
	DestinationID string      `json:"destination_id"`
	Highlights    JSONArr     `json:"highlights"`
	Badge         string      `json:"badge"`
	Badges        []string    `json:"badges"`
}

func (e *Event) ToResponse(trendingIDs map[string]bool) EventResponse {
	var badges []string

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
		ID:            e.ID,
		ExternalID:    e.ExternalID,
		Title:         e.Title,
		Description:   e.Description,
		Location:      e.Location,
		StartDate:     e.StartDate,
		EndDate:       e.EndDate,
		ImageURL:      e.ImageURL,
		Category:      e.Category,
		Status:        e.Status,
		Latitude:      e.Latitude,
		Longitude:     e.Longitude,
		MaxAttendees:  e.MaxAttendees,
		TicketPrice:   e.TicketPrice,
		Organizer:     e.Organizer,
		DestinationID: e.DestinationID,
		Highlights:    e.Highlights,
		Badge:         primaryBadge,
		Badges:        badges,
	}
}

// loadTrendingIDs reads the AI-selected trending IDs from Redis
func loadTrendingIDs(cacheStore cache.Store) map[string]bool {
	var ids []string
	ok, err := cacheStore.GetJSON(context.Background(), cache.KeyAITrendingIDs, &ids)
	if err != nil || !ok || len(ids) == 0 {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}
