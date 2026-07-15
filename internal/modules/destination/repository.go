package destination

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Destination, error)
	FindByID(externalID string) (*Destination, error)
	FindBySlug(slug string) (*Destination, error)
	FindByCategory(category string) ([]Destination, error)
	Search(query string) ([]Destination, error)
	Create(dest *Destination) error
	CreateBatch(dests []Destination) error
	Update(dest *Destination) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Destination, error) {
	var dests []Destination
	err := r.db.Order("id ASC").Find(&dests).Error
	return dests, err
}

func (r *GormRepository) FindByID(externalID string) (*Destination, error) {
	var dest Destination
	err := r.db.Where("external_id = ?", externalID).First(&dest).Error
	if err != nil {
		// Fallback: try slug-based lookup (slugify name and match)
		return r.FindBySlug(externalID)
	}
	return &dest, nil
}

// FindBySlug looks up a destination by converting name to a URL slug and comparing.
// e.g. "Malioboro Street" → "malioboro-street"
func (r *GormRepository) FindBySlug(slug string) (*Destination, error) {
	var dest Destination
	err := r.db.Where(
		"LOWER(REGEXP_REPLACE(name, '[^a-zA-Z0-9]+', '-', 'g')) = ?",
		slug,
	).First(&dest).Error
	if err != nil {
		return nil, err
	}
	return &dest, nil
}

func (r *GormRepository) FindByCategory(category string) ([]Destination, error) {
	var dests []Destination
	err := r.db.Where("category = ?", category).Order("rating DESC").Find(&dests).Error
	return dests, err
}

func (r *GormRepository) Search(query string) ([]Destination, error) {
	var dests []Destination
	like := "%" + query + "%"
	err := r.db.Where(
		r.db.Where("name ILIKE ?", like).
			Or("tagline ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("location ILIKE ?", like).
			Or("category ILIKE ?", like).
			Or("sub_region ILIKE ?", like),
	).Order("rating DESC").Find(&dests).Error
	return dests, err
}

func (r *GormRepository) Create(dest *Destination) error {
	return r.db.Create(dest).Error
}

func (r *GormRepository) CreateBatch(dests []Destination) error {
	return r.db.CreateInBatches(dests, 50).Error
}

func (r *GormRepository) Update(dest *Destination) error {
	return r.db.Save(dest).Error
}
