package restaurant

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Restaurant, error)
	FindByID(externalID string) (*Restaurant, error)
	Search(query string) ([]Restaurant, error)
	Create(restaurant *Restaurant) error
	Update(restaurant *Restaurant) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Restaurant, error) {
	var restaurants []Restaurant
	err := r.db.Order("id ASC").Find(&restaurants).Error
	return restaurants, err
}

func (r *GormRepository) FindByID(externalID string) (*Restaurant, error) {
	var restaurant Restaurant
	err := r.db.Where("external_id = ?", externalID).First(&restaurant).Error
	if err != nil {
		return nil, err
	}
	return &restaurant, nil
}

func (r *GormRepository) Search(query string) ([]Restaurant, error) {
	var restaurants []Restaurant
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like).
			Or("cuisine_type ILIKE ?", like),
	).Order("rating DESC").Find(&restaurants).Error
	return restaurants, err
}

func (r *GormRepository) Create(restaurant *Restaurant) error {
	return r.db.Create(restaurant).Error
}

func (r *GormRepository) Update(restaurant *Restaurant) error {
	return r.db.Save(restaurant).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Restaurant{}).Error
}
