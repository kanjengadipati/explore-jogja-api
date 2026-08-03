package subscription

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindByExternalID(externalID string) (*Subscription, error)
	FindByBusinessID(businessID uint) (*Subscription, error)
	FindByBusinessExternalID(businessExternalID string) (*Subscription, error)
	Create(subscription *Subscription) error
	Update(subscription *Subscription) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByExternalID(externalID string) (*Subscription, error) {
	var sub Subscription
	err := r.db.Where("external_id = ?", externalID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *GormRepository) FindByBusinessID(businessID uint) (*Subscription, error) {
	var sub Subscription
	err := r.db.Where("business_id = ?", businessID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *GormRepository) FindByBusinessExternalID(businessExternalID string) (*Subscription, error) {
	var sub Subscription
	err := r.db.Table("subscriptions").
		Joins("JOIN businesses ON businesses.id = subscriptions.business_id").
		Where("businesses.external_id = ?", businessExternalID).
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *GormRepository) Create(subscription *Subscription) error {
	return r.db.Create(subscription).Error
}

func (r *GormRepository) Update(subscription *Subscription) error {
	return r.db.Save(subscription).Error
}
