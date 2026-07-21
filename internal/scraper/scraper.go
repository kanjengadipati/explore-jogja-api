package scraper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// Registry holds all registered scrapers.
var scrapers []Scraper

// Register adds a scraper to the registry.
func Register(s Scraper) {
	scrapers = append(scrapers, s)
}

// RunAll runs all registered scrapers and persists results to the database.
func RunAll(db *gorm.DB) []ScrapeResult {
	destMap := buildDestMap(db)
	var results []ScrapeResult
	for _, s := range scrapers {
		log.Printf("[scraper] starting %s", s.Name())
		result := ScrapeResult{Source: s.Name()}

		dests, err := s.ScrapeDestinations()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("destinations: %v", err))
			log.Printf("[scraper] %s destinations error: %v", s.Name(), err)
		} else {
			di, du := upsertDestinations(db, dests, s.Name())
			result.DestinationsInserted = di
			result.DestinationsUpdated = du
			log.Printf("[scraper] %s destinations: %d inserted, %d updated", s.Name(), di, du)
		}

		events, err := s.ScrapeEvents()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("events: %v", err))
			log.Printf("[scraper] %s events error: %v", s.Name(), err)
		} else {
			ei, eu := upsertEvents(db, events, s.Name(), destMap)
			result.EventsInserted = ei
			result.EventsUpdated = eu
			log.Printf("[scraper] %s events: %d inserted, %d updated", s.Name(), ei, eu)
		}

		results = append(results, result)
	}
	return results
}

// buildDestMap loads all destination external_id → name for matching.
func buildDestMap(db *gorm.DB) map[string]string {
	var dests []destination.Destination
	db.Select("external_id, name").Find(&dests)
	m := make(map[string]string, len(dests))
	for _, d := range dests {
		m[d.ExternalID] = strings.ToLower(d.Name)
	}
	return m
}

func upsertDestinations(db *gorm.DB, items []ScrapedDestination, source string) (int, int) {
	inserted := 0
	updated := 0
	now := time.Now()

	for _, item := range items {
		if item.ExternalID == "" {
			item.ExternalID = slug.Make(item.Name)
		}

		var existing destination.Destination
		err := db.Where("external_id = ?", item.ExternalID).First(&existing).Error
		if err != nil {
			d := destination.Destination{
				ExternalID:  item.ExternalID,
				Name:        item.Name,
				Tagline:     item.Tagline,
				Category:    item.Category,
				Location:    item.Location,
				SubRegion:   item.SubRegion,
				Images:      strsToDestJSONArr(item.Images),
				Description: item.Description,
				Story:       item.Story,
				TicketPrice: item.TicketPrice,
				Latitude:    item.Latitude,
				Longitude:   item.Longitude,
			}
			if err := db.Create(&d).Error; err != nil {
				log.Printf("[scraper] failed to create destination %s: %v", item.ExternalID, err)
				continue
			}
			inserted++
			continue
		}

		if existing.UpdatedAt.Before(now) {
			existing.Name = item.Name
			existing.Tagline = item.Tagline
			existing.Category = item.Category
			existing.Location = item.Location
			existing.SubRegion = item.SubRegion
			if len(item.Images) > 0 {
				existing.Images = strsToDestJSONArr(item.Images)
			}
			existing.Description = item.Description
			existing.Story = item.Story
			existing.TicketPrice = item.TicketPrice
			existing.Latitude = item.Latitude
			existing.Longitude = item.Longitude
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("[scraper] failed to update destination %s: %v", item.ExternalID, err)
				continue
			}
			updated++
		}
	}
	return inserted, updated
}

func upsertEvents(db *gorm.DB, items []ScrapedEvent, source string, destMap map[string]string) (int, int) {
	inserted := 0
	updated := 0
	now := time.Now()

	for _, item := range items {
		if item.ExternalID == "" {
			item.ExternalID = slug.Make(item.Title)
		}

		// auto-match destination if not already set
		if item.DestinationID == "" {
			item.DestinationID = matchFromDestMap(destMap, item.Title, item.Location)
		}

		var existing event.Event
		err := db.Where("external_id = ?", item.ExternalID).First(&existing).Error
		if err != nil {
			e := event.Event{
				ExternalID:    item.ExternalID,
				Title:         item.Title,
				Description:   item.Description,
				Location:      item.Location,
				StartDate:     item.StartDate,
				EndDate:       item.EndDate,
				ImageURL:      item.ImageURL,
				Category:      item.Category,
				Status:        "upcoming",
				Latitude:      item.Latitude,
				Longitude:     item.Longitude,
				TicketPrice:   item.TicketPrice,
				Organizer:     item.Organizer,
				Highlights:    strsToEventJSONArr(item.Highlights),
				DestinationID: item.DestinationID,
			}
			if err := db.Create(&e).Error; err != nil {
				log.Printf("[scraper] failed to create event %s: %v", item.ExternalID, err)
				continue
			}
			inserted++
			continue
		}

		if existing.UpdatedAt.Before(now) {
			existing.Title = item.Title
			existing.Description = item.Description
			existing.Location = item.Location
			existing.StartDate = item.StartDate
			existing.EndDate = item.EndDate
			if item.ImageURL != "" {
				existing.ImageURL = item.ImageURL
			}
			existing.Category = item.Category
			existing.Latitude = item.Latitude
			existing.Longitude = item.Longitude
			existing.TicketPrice = item.TicketPrice
			existing.Organizer = item.Organizer
			if len(item.Highlights) > 0 {
				existing.Highlights = strsToEventJSONArr(item.Highlights)
			}
			if item.DestinationID != "" {
				existing.DestinationID = item.DestinationID
			}
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("[scraper] failed to update event %s: %v", item.ExternalID, err)
				continue
			}
			updated++
		}
	}
	return inserted, updated
}

func strsToDestJSONArr(s []string) destination.JSONArr {
	arr := make(destination.JSONArr, len(s))
	for i, v := range s {
		arr[i] = v
	}
	return arr
}

func strsToEventJSONArr(s []string) event.JSONArr {
	arr := make(event.JSONArr, len(s))
	for i, v := range s {
		arr[i] = v
	}
	return arr
}

func slugify(s string) string {
	return slug.Make(s)
}

func imgs(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// matchFromDestMap tries to find a destination whose name is contained in the event
// title or location.
func matchFromDestMap(destMap map[string]string, title, location string) string {
	titleLower := strings.ToLower(title)
	locLower := strings.ToLower(location)

	for extID, nameLower := range destMap {
		if nameLower == "" {
			continue
		}
		if strings.Contains(titleLower, nameLower) {
			return extID
		}
		if locLower != "" && strings.Contains(locLower, nameLower) {
			return extID
		}
	}

	return ""
}
