package scraper

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// StartScheduler starts a cron job that runs all scrapers on the given schedule.
// Default: every Sunday at 00:00 (midnight).
// Pass an empty string to use the default schedule.
func StartScheduler(db *gorm.DB, schedule string) {
	if schedule == "" {
		schedule = "0 0 * * 0" // every Sunday at midnight
	}

	c := cron.New()
	_, err := c.AddFunc(schedule, func() {
		log.Printf("[scraper] scheduled run started at %s", time.Now().Format(time.RFC3339))
		results := RunAll(db)
		for _, r := range results {
			log.Printf("[scraper] %s complete: events(%d/%d) destinations(%d/%d) errors(%d)",
				r.Source, r.EventsInserted, r.EventsUpdated,
				r.DestinationsInserted, r.DestinationsUpdated,
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
