package story

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Story, error)
	FindByID(externalID string) (*Story, error)
	Search(query string) ([]Story, error)
	Create(story *Story) error
	Update(story *Story) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Story, error) {
	var stories []Story
	err := r.db.Order("id ASC").Find(&stories).Error
	return stories, err
}

func (r *GormRepository) FindByID(externalID string) (*Story, error) {
	var story Story
	err := r.db.Where("external_id = ?", externalID).First(&story).Error
	if err != nil {
		return nil, err
	}
	return &story, nil
}

func (r *GormRepository) Search(query string) ([]Story, error) {
	var stories []Story
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("title ILIKE ?", like).
			Or("content ILIKE ?", like),
	).Order("created_at DESC").Find(&stories).Error
	return stories, err
}

func (r *GormRepository) Create(story *Story) error {
	return r.db.Create(story).Error
}

func (r *GormRepository) Update(story *Story) error {
	return r.db.Save(story).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Story{}).Error
}
