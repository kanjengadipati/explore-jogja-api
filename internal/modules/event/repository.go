package event

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Event, error)
	FindByID(externalID string) (*Event, error)
	Search(query string) ([]Event, error)
	Create(event *Event) error
	Update(event *Event) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Event, error) {
	var events []Event
	err := r.db.Order("id ASC").Find(&events).Error
	return events, err
}

func (r *GormRepository) FindByID(externalID string) (*Event, error) {
	var event Event
	err := r.db.Where("external_id = ?", externalID).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *GormRepository) Search(query string) ([]Event, error) {
	var events []Event
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("title ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like).
			Or("category ILIKE ?", like),
	).Order("start_date DESC").Find(&events).Error
	return events, err
}

func (r *GormRepository) Create(event *Event) error {
	return r.db.Create(event).Error
}

func (r *GormRepository) Update(event *Event) error {
	return r.db.Save(event).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Event{}).Error
}
