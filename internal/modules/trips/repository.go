package trips

import "gorm.io/gorm"

type Repository interface {
	FindByUser(userID uint) ([]Trip, error)
	FindByID(externalID string, userID uint) (*Trip, error)
	Create(trip *Trip) error
	Update(trip *Trip) error
	Delete(externalID string, userID uint) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByUser(userID uint) ([]Trip, error) {
	var trips []Trip
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&trips).Error
	return trips, err
}

func (r *GormRepository) FindByID(externalID string, userID uint) (*Trip, error) {
	var trip Trip
	err := r.db.Where("external_id = ? AND user_id = ?", externalID, userID).
		First(&trip).Error
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *GormRepository) Create(trip *Trip) error {
	return r.db.Create(trip).Error
}

func (r *GormRepository) Update(trip *Trip) error {
	return r.db.Save(trip).Error
}

func (r *GormRepository) Delete(externalID string, userID uint) error {
	return r.db.Where("external_id = ? AND user_id = ?", externalID, userID).
		Delete(&Trip{}).Error
}
