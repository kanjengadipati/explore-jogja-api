package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"
	"pleco-api/internal/modules/staging"

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
			di, du, ds := upsertDestinations(db, dests, s.Name())
			result.DestinationsInserted = di
			result.DestinationsUpdated = du
			result.DestinationsStaged = ds
			log.Printf("[scraper] %s destinations: %d inserted, %d updated, %d staged", s.Name(), di, du, ds)
		}

		events, err := s.ScrapeEvents()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("events: %v", err))
			log.Printf("[scraper] %s events error: %v", s.Name(), err)
		} else {
			ei, eu, es := upsertEvents(db, events, s.Name(), destMap)
			result.EventsInserted = ei
			result.EventsUpdated = eu
			result.EventsStaged = es
			log.Printf("[scraper] %s events: %d inserted, %d updated, %d staged", s.Name(), ei, eu, es)
		}

		results = append(results, result)
	}

	// Fix any stale event statuses based on dates
	BackfillEventStatuses(db)

	// Populate missing videos for a few items per run
	populateMissingVideos(db)

	return results
}

// RunDestinationsOnly runs only destination scrapers (for monthly schedule).
func RunDestinationsOnly(db *gorm.DB) []ScrapeResult {
	var results []ScrapeResult
	for _, s := range scrapers {
		log.Printf("[scraper] starting %s destinations", s.Name())
		result := ScrapeResult{Source: s.Name()}

		dests, err := s.ScrapeDestinations()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("destinations: %v", err))
			log.Printf("[scraper] %s destinations error: %v", s.Name(), err)
		} else {
			di, du, ds := upsertDestinations(db, dests, s.Name())
			result.DestinationsInserted = di
			result.DestinationsUpdated = du
			result.DestinationsStaged = ds
			log.Printf("[scraper] %s destinations: %d inserted, %d updated, %d staged", s.Name(), di, du, ds)
		}

		results = append(results, result)
	}

	// Populate missing videos for destinations only
	var dests []destination.Destination
	db.Where("video_url = '' OR video_url IS NULL").Limit(10).Find(&dests)
	for _, d := range dests {
		url := FetchYouTubeVideoURL(d.Name)
		if url != "" {
			db.Model(&d).Update("video_url", url)
			log.Printf("[scraper] auto-populated video for destination: %s", d.Name)
		}
	}

	return results
}

// RunEventsOnly runs only event scrapers (for 3-day schedule).
func RunEventsOnly(db *gorm.DB) []ScrapeResult {
	destMap := buildDestMap(db)
	var results []ScrapeResult
	for _, s := range scrapers {
		log.Printf("[scraper] starting %s events", s.Name())
		result := ScrapeResult{Source: s.Name()}

		events, err := s.ScrapeEvents()
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("events: %v", err))
			log.Printf("[scraper] %s events error: %v", s.Name(), err)
		} else {
			ei, eu, es := upsertEvents(db, events, s.Name(), destMap)
			result.EventsInserted = ei
			result.EventsUpdated = eu
			result.EventsStaged = es
			log.Printf("[scraper] %s events: %d inserted, %d updated, %d staged", s.Name(), ei, eu, es)
		}

		results = append(results, result)
	}

	// Populate missing videos for events only
	var events []event.Event
	db.Where("video_url = '' OR video_url IS NULL").Limit(10).Find(&events)
	for _, e := range events {
		url := FetchYouTubeVideoURL(e.Title + " " + e.Location)
		if url != "" {
			db.Model(&e).Update("video_url", url)
			log.Printf("[scraper] auto-populated video for event: %s", e.Title)
		}
	}

	return results
}

func populateMissingVideos(db *gorm.DB) {
	// Limit to 10 each to save quota
	var dests []destination.Destination
	db.Where("video_url = '' OR video_url IS NULL").Limit(10).Find(&dests)
	for _, d := range dests {
		url := FetchYouTubeVideoURL(d.Name)
		if url != "" {
			db.Model(&d).Update("video_url", url)
			log.Printf("[scraper] auto-populated video for destination: %s", d.Name)
		}
	}

	var events []event.Event
	db.Where("video_url = '' OR video_url IS NULL").Limit(10).Find(&events)
	for _, e := range events {
		url := FetchYouTubeVideoURL(e.Title + " " + e.Location)
		if url != "" {
			db.Model(&e).Update("video_url", url)
			log.Printf("[scraper] auto-populated video for event: %s", e.Title)
		}
	}
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

func upsertDestinations(db *gorm.DB, items []ScrapedDestination, source string) (int, int, int) {
	inserted := 0
	updated := 0
	staged := 0
	now := time.Now()

	for _, item := range items {
		if item.ExternalID == "" {
			item.ExternalID = slug.Make(item.Name)
		}

		var existing destination.Destination
		err := db.Where("external_id = ?", item.ExternalID).First(&existing).Error
		if err != nil {
			// Not published yet → send to staging for review/approval instead
			// of inserting straight into the live table.
			stageDestination(db, item, source)
			staged++
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
			if item.VideoURL != "" {
				existing.VideoURL = item.VideoURL
			}
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("[scraper] failed to update destination %s: %v", item.ExternalID, err)
				continue
			}
			updated++
		}
	}
	return inserted, updated, staged
}

func upsertEvents(db *gorm.DB, items []ScrapedEvent, source string, destMap map[string]string) (int, int, int) {
	inserted := 0
	updated := 0
	staged := 0
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
			// Not published yet → send to staging for review/approval.
			stageEvent(db, item, source)
			staged++
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
			if item.VideoURL != "" {
				existing.VideoURL = item.VideoURL
			}
			existing.Status = resolveEventStatus(item.StartDate, item.EndDate)
			if err := db.Save(&existing).Error; err != nil {
				log.Printf("[scraper] failed to update event %s: %v", item.ExternalID, err)
				continue
			}
			updated++
		}
	}
	return inserted, updated, staged
}

// stageDestination queues a brand-new destination for human review. Already
// queued providers (any status) are left untouched so re-runs don't duplicate
// or resurrect rejected items.
func stageDestination(db *gorm.DB, item ScrapedDestination, source string) {
	var existing staging.StagingDestination
	err := db.Where("provider_id = ? AND source = ?", item.ExternalID, source).First(&existing).Error
	if err == nil {
		return
	}

	images, _ := json.Marshal(item.Images)
	raw, _ := json.Marshal(item)

	lat, lng := "", ""
	if item.Latitude != 0 {
		lat = strconv.FormatFloat(item.Latitude, 'f', -1, 64)
	}
	if item.Longitude != 0 {
		lng = strconv.FormatFloat(item.Longitude, 'f', -1, 64)
	}

	row := staging.StagingDestination{
		Source:      source,
		ProviderID:  item.ExternalID,
		Name:        item.Name,
		Description: item.Description,
		Latitude:    lat,
		Longitude:   lng,
		Address:     item.Location,
		Category:    item.Category,
		Images:      string(images),
		RawData:     string(raw),
		Status:      "pending",
	}
	if err := db.Create(&row).Error; err != nil {
		log.Printf("[scraper] failed to stage destination %s: %v", item.ExternalID, err)
	}
}

// stageEvent queues a brand-new event for human review.
func stageEvent(db *gorm.DB, item ScrapedEvent, source string) {
	var existing staging.StagingEvent
	err := db.Where("provider_id = ? AND source = ?", item.ExternalID, source).First(&existing).Error
	if err == nil {
		return
	}

	raw, _ := json.Marshal(item)

	var start, end time.Time
	if t, err := time.Parse("2006-01-02", item.StartDate); err == nil {
		start = t
	}
	if t, err := time.Parse("2006-01-02", item.EndDate); err == nil {
		end = t
	}

	row := staging.StagingEvent{
		Source:      source,
		ProviderID:  item.ExternalID,
		Title:       item.Title,
		Description: item.Description,
		StartDate:   start,
		EndDate:     end,
		Location:    item.Location,
		RawData:     string(raw),
		Status:      "pending",
	}
	if err := db.Create(&row).Error; err != nil {
		log.Printf("[scraper] failed to stage event %s: %v", item.ExternalID, err)
	}
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

// resolveEventStatus determines the event status based on start/end dates.
func resolveEventStatus(startDate, endDate string) string {
	now := time.Now()
	_layout := "2006-01-02"

	if startDate != "" {
		if start, err := time.Parse(_layout, startDate); err == nil && start.After(now) {
			return "upcoming"
		}
	}
	if endDate != "" {
		if end, err := time.Parse(_layout, endDate); err == nil && end.Before(now) {
			return "completed"
		}
	}
	if startDate != "" {
		if _, err := time.Parse(_layout, startDate); err == nil {
			return "active"
		}
	}
	return "upcoming"
}

// BackfillEventStatuses fixes statuses for all events in the DB based on their dates.
func BackfillEventStatuses(db *gorm.DB) {
	var events []event.Event
	db.Find(&events)
	updated := 0
	for _, e := range events {
		correct := resolveEventStatus(e.StartDate, e.EndDate)
		if correct != e.Status {
			db.Model(&e).Update("status", correct)
			updated++
		}
	}
	if updated > 0 {
		log.Printf("[scraper] backfilled status for %d events", updated)
	}
}

func imgs(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

// normalizeSubRegion normalizes raw location text into a standard sub_region value.
// e.g. "Kabupaten Bantul" → "Bantul", "Kota Yogyakarta" → "Yogyakarta".
func normalizeSubRegion(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	// strip common prefixes
	s = strings.TrimPrefix(s, "Kabupaten ")
	s = strings.TrimPrefix(s, "Kab. ")
	s = strings.TrimPrefix(s, "Kota ")
	s = strings.TrimPrefix(s, "Kota ")
	s = strings.TrimSpace(s)

	// map known variations to canonical names
	lower := strings.ToLower(s)
	regionMap := map[string]string{
		"bantul":       "Bantul",
		"gunungkidul":  "Gunungkidul",
		"gunung kidul": "Gunungkidul",
		"kulon progo":  "Kulon Progo",
		"sleman":       "Sleman",
		"yogyakarta":   "Yogyakarta",
		"jogja":        "Yogyakarta",
		"jogjakarta":   "Yogyakarta",
	}
	if canonical, ok := regionMap[lower]; ok {
		return canonical
	}
	// fallback: title-case the cleaned string
	if len(s) > 0 {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	return s
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
