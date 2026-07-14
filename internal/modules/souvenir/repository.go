package souvenir

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Souvenir, error)
	FindByID(externalID string) (*Souvenir, error)
	Search(query string) ([]Souvenir, error)
	Create(souvenir *Souvenir) error
	Update(souvenir *Souvenir) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Souvenir, error) {
	var souvenirs []Souvenir
	err := r.db.Order("id ASC").Find(&souvenirs).Error
	return souvenirs, err
}

func (r *GormRepository) FindByID(externalID string) (*Souvenir, error) {
	var souvenir Souvenir
	err := r.db.Where("external_id = ?", externalID).First(&souvenir).Error
	if err != nil {
		return nil, err
	}
	return &souvenir, nil
}

func (r *GormRepository) Search(query string) ([]Souvenir, error) {
	var souvenirs []Souvenir
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like),
	).Order("rating DESC").Find(&souvenirs).Error
	return souvenirs, err
}

func (r *GormRepository) Create(souvenir *Souvenir) error {
	return r.db.Create(souvenir).Error
}

func (r *GormRepository) Update(souvenir *Souvenir) error {
	return r.db.Save(souvenir).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Souvenir{}).Error
}
