package staging

import (
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
	return r.db.Model(&StagingDestination{}).Where("id = ?", id).Update("status", "approved").Error
}

func (r *gormRepository) RejectDestination(id uint) error {
	return r.db.Model(&StagingDestination{}).Where("id = ?", id).Update("status", "rejected").Error
}

func (r *gormRepository) ApproveMultipleDestinations(ids []uint) error {
	return r.db.Model(&StagingDestination{}).Where("id IN ?", ids).Update("status", "approved").Error
}

func (r *gormRepository) RejectMultipleDestinations(ids []uint) error {
	return r.db.Model(&StagingDestination{}).Where("id IN ?", ids).Update("status", "rejected").Error
}

func (r *gormRepository) ApproveMultipleEvents(ids []uint) error {
	return r.db.Model(&StagingEvent{}).Where("id IN ?", ids).Update("status", "approved").Error
}

func (r *gormRepository) RejectMultipleEvents(ids []uint) error {
	return r.db.Model(&StagingEvent{}).Where("id IN ?", ids).Update("status", "rejected").Error
}
