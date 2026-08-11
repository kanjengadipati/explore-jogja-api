package scraper

import (
	"log"
	"sync"
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

// StartSplitScheduler starts two separate jobs:
//   - destinations: scrapes destinations only (default: 1st of every month at 01:00)
//   - events: scrapes events only on a rolling 3-day cadence (daily tick that
//     runs at most once per 72 hours, avoiding the 1-day drift a "*/3"
//     day-of-month cron would introduce at month boundaries)
func StartSplitScheduler(db *gorm.DB, destSchedule, eventSchedule string) {
	if destSchedule == "" {
		destSchedule = "0 0 1 * *" // 1st of every month at midnight
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

	// Events — daily tick gated by a rolling interval. A raw "*/3" day-of-month
	// cron resets at each month boundary; this keeps a true 3-day spacing.
	if eventSchedule != "" && eventSchedule != "0 0 */3 * *" {
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
	} else {
		var (
			lastEventRunMu sync.Mutex
			lastEventRun   time.Time
		)
		_, err = c.AddFunc("0 0 * * *", func() {
			lastEventRunMu.Lock()
			defer lastEventRunMu.Unlock()
			if !lastEventRun.IsZero() && time.Since(lastEventRun) < 3*24*time.Hour {
				return
			}
			lastEventRun = time.Now()
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
	}

	c.Start()
	log.Printf("[scraper] split scheduler started — destinations: %s, events: rolling every 3 days", destSchedule)
}
