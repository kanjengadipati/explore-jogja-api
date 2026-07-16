package config

import "gorm.io/gorm"

type Repository interface {
	GetAll() ([]SiteConfig, error)
	GetByKey(key string) (*SiteConfig, error)
	GetByCategory(category string) ([]SiteConfig, error)
	Upsert(key, value, category string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) GetAll() ([]SiteConfig, error) {
	var configs []SiteConfig
	err := r.db.Order("category ASC, key ASC").Find(&configs).Error
	return configs, err
}

func (r *GormRepository) GetByKey(key string) (*SiteConfig, error) {
	var config SiteConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *GormRepository) GetByCategory(category string) ([]SiteConfig, error) {
	var configs []SiteConfig
	err := r.db.Where("category = ?", category).Order("key ASC").Find(&configs).Error
	return configs, err
}

func (r *GormRepository) Upsert(key, value, category string) error {
	var existing SiteConfig
	err := r.db.Where("key = ?", key).First(&existing).Error
	if err == nil {
		existing.Value = value
		if category != "" {
			existing.Category = category
		}
		return r.db.Save(&existing).Error
	}
	config := SiteConfig{
		Key:      key,
		Value:    value,
		Category: category,
	}
	return r.db.Create(&config).Error
}
