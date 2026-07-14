package hotel

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Hotel, error)
	FindByID(externalID string) (*Hotel, error)
	Search(query string) ([]Hotel, error)
	Create(hotel *Hotel) error
	Update(hotel *Hotel) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Hotel, error) {
	var hotels []Hotel
	err := r.db.Order("id ASC").Find(&hotels).Error
	return hotels, err
}

func (r *GormRepository) FindByID(externalID string) (*Hotel, error) {
	var hotel Hotel
	err := r.db.Where("external_id = ?", externalID).First(&hotel).Error
	if err != nil {
		return nil, err
	}
	return &hotel, nil
}

func (r *GormRepository) Search(query string) ([]Hotel, error) {
	var hotels []Hotel
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like).
			Or("address ILIKE ?", like),
	).Order("rating DESC").Find(&hotels).Error
	return hotels, err
}

func (r *GormRepository) Create(hotel *Hotel) error {
	return r.db.Create(hotel).Error
}

func (r *GormRepository) Update(hotel *Hotel) error {
	return r.db.Save(hotel).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Hotel{}).Error
}
