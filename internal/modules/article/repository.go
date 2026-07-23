package article

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(status string) ([]Article, error)
	FindBySlug(slug string) (*Article, error)
	FindByID(externalID string) (*Article, error)
	FindByCategory(category string) ([]Article, error)
	Search(query string) ([]Article, error)
	Create(a *Article) error
	CreateBatch(articles []Article) error
	Update(a *Article) error
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll(status string) ([]Article, error) {
	var articles []Article
	q := r.db.Order("published_at DESC, created_at DESC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&articles).Error
	return articles, err
}

func (r *GormRepository) FindBySlug(slug string) (*Article, error) {
	var a Article
	err := r.db.Where("slug = ?", slug).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GormRepository) FindByID(externalID string) (*Article, error) {
	var a Article
	err := r.db.Where("external_id = ?", externalID).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GormRepository) FindByCategory(category string) ([]Article, error) {
	var articles []Article
	err := r.db.Where("category = ? AND status = ?", category, "published").
		Order("published_at DESC").Find(&articles).Error
	return articles, err
}

func (r *GormRepository) Search(query string) ([]Article, error) {
	var articles []Article
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("title ILIKE ?", like).
			Or("title_en ILIKE ?", like).
			Or("excerpt ILIKE ?", like).
			Or("content ILIKE ?", like).
			Or("category ILIKE ?", like),
	).Where("status = ?", "published").
		Order("published_at DESC").Find(&articles).Error
	return articles, err
}

func (r *GormRepository) Create(a *Article) error {
	return r.db.Create(a).Error
}

func (r *GormRepository) CreateBatch(articles []Article) error {
	return r.db.CreateInBatches(articles, 20).Error
}

func (r *GormRepository) Update(a *Article) error {
	return r.db.Save(a).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Article{}).Error
}
