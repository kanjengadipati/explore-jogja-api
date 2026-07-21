package sitemap

import (
	"fmt"
	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	baseURL string
}

func NewHandler(db *gorm.DB, baseURL string) *Handler {
	return &Handler{db: db, baseURL: baseURL}
}

func (h *Handler) GetSitemap(c *gin.Context) {
	var dests []destination.Destination
	h.db.Select("external_id, updated_at").Find(&dests)

	var events []event.Event
	h.db.Select("external_id, updated_at").Find(&events)

	c.Header("Content-Type", "application/xml")
	c.Writer.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	c.Writer.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// Home
	writeURL(c, h.baseURL, time.Now())

	// Destinations
	for _, d := range dests {
		writeURL(c, fmt.Sprintf("%s/destinations/%s", h.baseURL, d.ExternalID), d.UpdatedAt)
	}

	// Events
	for _, e := range events {
		writeURL(c, fmt.Sprintf("%s/events/%s", h.baseURL, e.ExternalID), e.UpdatedAt)
	}

	c.Writer.WriteString(`</urlset>`)
}

func writeURL(c *gin.Context, loc string, lastmod time.Time) {
	c.Writer.WriteString(fmt.Sprintf("  <url>\n    <loc>%s</loc>\n    <lastmod>%s</lastmod>\n  </url>\n", loc, lastmod.Format("2006-01-02")))
}
