package partner

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Partner, error)
	FindByID(externalID string) (*Partner, error)
	Search(query string) ([]Partner, error)
	Create(partner *Partner) error
	Update(partner *Partner) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Partner, error) {
	var partners []Partner
	err := r.db.Order("id ASC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) FindByID(externalID string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("external_id = ?", externalID).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *GormRepository) Search(query string) ([]Partner, error) {
	var partners []Partner
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like).
			Or("category ILIKE ?", like),
	).Order("rating DESC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) Create(partner *Partner) error {
	return r.db.Create(partner).Error
}

func (r *GormRepository) Update(partner *Partner) error {
	return r.db.Save(partner).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Partner{}).Error
}
