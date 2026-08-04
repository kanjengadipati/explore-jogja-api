package scraper

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// StartScheduler starts a cron job that runs all scrapers on the given schedule.
// Deprecated: Use StartSplitScheduler for separate destination and event schedules.
func StartScheduler(db *gorm.DB, schedule string) {
	if schedule == "" {
		schedule = "0 0 * * 0"
	}

	c := cron.New()
	_, err := c.AddFunc(schedule, func() {
		log.Printf("[scraper] scheduled run started at %s", time.Now().Format(time.RFC3339))
		results := RunAll(db)
		for _, r := range results {
			log.Printf("[scraper] %s complete: events(upd=%d staged=%d) destinations(upd=%d staged=%d) errors(%d)",
				r.Source, r.EventsUpdated, r.EventsStaged,
				r.DestinationsUpdated, r.DestinationsStaged,
				len(r.Errors))
		}
		log.Printf("[scraper] scheduled run finished at %s", time.Now().Format(time.RFC3339))
	})
	if err != nil {
		log.Fatalf("[scraper] failed to register cron job: %v", err)
	}

	c.Start()
	log.Printf("[scraper] scheduler started with schedule: %s", schedule)
}

// StartSplitScheduler starts two separate cron jobs:
//   - destinations: scrapes destinations only (default: 1st of every month at 01:00)
//   - events: scrapes events only (default: every 3 days at 00:00)
func StartSplitScheduler(db *gorm.DB, destSchedule, eventSchedule string) {
	if destSchedule == "" {
		destSchedule = "0 0 1 * *" // 1st of every month at midnight
	}
	if eventSchedule == "" {
		eventSchedule = "0 0 */3 * *" // every 3 days at midnight
	}

	c := cron.New()

	// Destinations cron
	_, err := c.AddFunc(destSchedule, func() {
		log.Printf("[scraper] destinations scheduled run started at %s", time.Now().Format(time.RFC3339))
		results := RunDestinationsOnly(db)
		for _, r := range results {
			log.Printf("[scraper] %s destinations complete: updated(%d) staged(%d) errors(%d)",
				r.Source, r.DestinationsUpdated, r.DestinationsStaged, len(r.Errors))
		}
		log.Printf("[scraper] destinations scheduled run finished at %s", time.Now().Format(time.RFC3339))
	})
	if err != nil {
		log.Fatalf("[scraper] failed to register destinations cron job: %v", err)
	}

	// Events cron
	_, err = c.AddFunc(eventSchedule, func() {
		log.Printf("[scraper] events scheduled run started at %s", time.Now().Format(time.RFC3339))
		results := RunEventsOnly(db)
		for _, r := range results {
			log.Printf("[scraper] %s events complete: updated(%d) staged(%d) errors(%d)",
				r.Source, r.EventsUpdated, r.EventsStaged, len(r.Errors))
		}
		log.Printf("[scraper] events scheduled run finished at %s", time.Now().Format(time.RFC3339))
	})
	if err != nil {
		log.Fatalf("[scraper] failed to register events cron job: %v", err)
	}

	c.Start()
	log.Printf("[scraper] split scheduler started — destinations: %s, events: %s", destSchedule, eventSchedule)
}
