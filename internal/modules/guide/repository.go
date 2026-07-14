package guide

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Guide, error)
	FindByID(externalID string) (*Guide, error)
	Search(query string) ([]Guide, error)
	Create(guide *Guide) error
	Update(guide *Guide) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Guide, error) {
	var guides []Guide
	err := r.db.Order("id ASC").Find(&guides).Error
	return guides, err
}

func (r *GormRepository) FindByID(externalID string) (*Guide, error) {
	var guide Guide
	err := r.db.Where("external_id = ?", externalID).First(&guide).Error
	if err != nil {
		return nil, err
	}
	return &guide, nil
}

func (r *GormRepository) Search(query string) ([]Guide, error) {
	var guides []Guide
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("bio ILIKE ?", like).
			Or("specialization ILIKE ?", like),
	).Order("rating DESC").Find(&guides).Error
	return guides, err
}

func (r *GormRepository) Create(guide *Guide) error {
	return r.db.Create(guide).Error
}

func (r *GormRepository) Update(guide *Guide) error {
	return r.db.Save(guide).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Guide{}).Error
}
