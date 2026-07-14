package rental

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Rental, error)
	FindByID(externalID string) (*Rental, error)
	Search(query string) ([]Rental, error)
	Create(rental *Rental) error
	Update(rental *Rental) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Rental, error) {
	var rentals []Rental
	err := r.db.Order("id ASC").Find(&rentals).Error
	return rentals, err
}

func (r *GormRepository) FindByID(externalID string) (*Rental, error) {
	var rental Rental
	err := r.db.Where("external_id = ?", externalID).First(&rental).Error
	if err != nil {
		return nil, err
	}
	return &rental, nil
}

func (r *GormRepository) Search(query string) ([]Rental, error) {
	var rentals []Rental
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like),
	).Order("rating DESC").Find(&rentals).Error
	return rentals, err
}

func (r *GormRepository) Create(rental *Rental) error {
	return r.db.Create(rental).Error
}

func (r *GormRepository) Update(rental *Rental) error {
	return r.db.Save(rental).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Rental{}).Error
}
