package staging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"pleco-api/internal/modules/destination"
	"pleco-api/internal/modules/event"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

type Repository interface {
	CreateDestination(dest *StagingDestination) error
	CreateEvent(event *StagingEvent) error
	FindPendingDestinations() ([]StagingDestination, error)
	FindPendingEvents() ([]StagingEvent, error)
	ApproveDestination(id uint) error
	RejectDestination(id uint) error
	ApproveMultipleDestinations(ids []uint) error
	RejectMultipleDestinations(ids []uint) error
	ApproveMultipleEvents(ids []uint) error
	RejectMultipleEvents(ids []uint) error
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) CreateDestination(dest *StagingDestination) error {
	return r.db.Create(dest).Error
}

func (r *gormRepository) CreateEvent(event *StagingEvent) error {
	return r.db.Create(event).Error
}

func (r *gormRepository) FindPendingDestinations() ([]StagingDestination, error) {
	var dests []StagingDestination
	err := r.db.Where("status = ?", "pending").Find(&dests).Error
	return dests, err
}

func (r *gormRepository) FindPendingEvents() ([]StagingEvent, error) {
	var events []StagingEvent
	err := r.db.Where("status = ?", "pending").Find(&events).Error
	return events, err
}

func (r *gormRepository) ApproveDestination(id uint) error {
	return r.ApproveMultipleDestinations([]uint{id})
}

func (r *gormRepository) RejectDestination(id uint) error {
	return r.db.Model(&StagingDestination{}).Where("id = ?", id).Update("status", "rejected").Error
}

func (r *gormRepository) RejectMultipleDestinations(ids []uint) error {
	return r.db.Model(&StagingDestination{}).Where("id IN ?", ids).Update("status", "rejected").Error
}

func (r *gormRepository) RejectMultipleEvents(ids []uint) error {
	return r.db.Model(&StagingEvent{}).Where("id IN ?", ids).Update("status", "rejected").Error
}

// ApproveMultipleDestinations publishes the approved staging rows into the
// live destinations table (upsert by external_id) and then marks them approved.
func (r *gormRepository) ApproveMultipleDestinations(ids []uint) error {
	var staged []StagingDestination
	if err := r.db.Where("id IN ?", ids).Find(&staged).Error; err != nil {
		return err
	}
	for _, s := range staged {
		if err := r.publishDestination(s); err != nil {
			log.Printf("[staging] failed to publish destination %s: %v", s.ProviderID, err)
		}
	}
	return r.db.Model(&StagingDestination{}).Where("id IN ?", ids).Update("status", "approved").Error
}

func (r *gormRepository) publishDestination(s StagingDestination) error {
	var images destination.JSONArr
	if s.Images != "" {
		_ = json.Unmarshal([]byte(s.Images), &images)
	}

	d := destination.Destination{
		ExternalID:  s.ProviderID,
		Name:        s.Name,
		Description: s.Description,
		Location:    s.Address,
		Category:    s.Category,
		Images:      images,
		Latitude:    parseFloat(s.Latitude),
		Longitude:   parseFloat(s.Longitude),
		Status:      "published",
	}
	if d.ExternalID == "" {
		d.ExternalID = slug.Make(d.Name)
	}
	if d.Name == "" {
		return nil
	}

	var existing destination.Destination
	err := r.db.Where("external_id = ?", d.ExternalID).First(&existing).Error
	if err != nil {
		return r.db.Create(&d).Error
	}
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"name":        d.Name,
		"description": d.Description,
		"location":    d.Location,
		"category":    d.Category,
		"images":      images,
		"latitude":    d.Latitude,
		"longitude":   d.Longitude,
		"status":      "published",
	}).Error
}

// ApproveMultipleEvents publishes the approved staging rows into the live
// events table (upsert by external_id) and then marks them approved.
func (r *gormRepository) ApproveMultipleEvents(ids []uint) error {
	var staged []StagingEvent
	if err := r.db.Where("id IN ?", ids).Find(&staged).Error; err != nil {
		return err
	}
	for _, s := range staged {
		if err := r.publishEvent(s); err != nil {
			log.Printf("[staging] failed to publish event %s: %v", s.ProviderID, err)
		}
	}
	return r.db.Model(&StagingEvent{}).Where("id IN ?", ids).Update("status", "approved").Error
}

func (r *gormRepository) publishEvent(s StagingEvent) error {
	startDate, endDate := "", ""
	if !s.StartDate.IsZero() {
		startDate = s.StartDate.Format("2006-01-02")
	}
	if !s.EndDate.IsZero() {
		endDate = s.EndDate.Format("2006-01-02")
	}

	e := event.Event{
		ExternalID:  s.ProviderID,
		Title:       s.Title,
		Description: s.Description,
		Location:    s.Location,
		StartDate:   startDate,
		EndDate:     endDate,
		Category:    "Event",
		Status:      resolveStagedStatus(startDate, endDate),
	}
	if e.ExternalID == "" {
		e.ExternalID = slug.Make(e.Title)
	}
	if e.Title == "" {
		return nil
	}

	var existing event.Event
	err := r.db.Where("external_id = ?", e.ExternalID).First(&existing).Error
	if err != nil {
		return r.db.Create(&e).Error
	}
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"title":       e.Title,
		"description": e.Description,
		"location":    e.Location,
		"start_date":  e.StartDate,
		"end_date":    e.EndDate,
		"category":    e.Category,
		"status":      e.Status,
	}).Error
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	var f float64
	if _, err := fmt.Sscan(s, &f); err != nil {
		return 0
	}
	return f
}

// resolveStagedStatus mirrors the scraper's status resolution so events
// published from staging get the correct status without importing the scraper
// package (which would create an import cycle).
func resolveStagedStatus(startDate, endDate string) string {
	now := time.Now()
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil && t.After(now) {
			return "upcoming"
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil && t.Before(now) {
			return "completed"
		}
	}
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err == nil {
			return "active"
		}
	}
	return "upcoming"
}
