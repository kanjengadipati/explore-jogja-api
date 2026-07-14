package review

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Review, error)
	FindByID(externalID string) (*Review, error)
	Search(query string) ([]Review, error)
	Create(review *Review) error
	Update(review *Review) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Review, error) {
	var reviews []Review
	err := r.db.Order("id ASC").Find(&reviews).Error
	return reviews, err
}

func (r *GormRepository) FindByID(externalID string) (*Review, error) {
	var review Review
	err := r.db.Where("external_id = ?", externalID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *GormRepository) Search(query string) ([]Review, error) {
	var reviews []Review
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("user_name ILIKE ?", like).
			Or("comment ILIKE ?", like),
	).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *GormRepository) Create(review *Review) error {
	return r.db.Create(review).Error
}

func (r *GormRepository) Update(review *Review) error {
	return r.db.Save(review).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Review{}).Error
}
