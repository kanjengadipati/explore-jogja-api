package destination

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(status string) ([]Destination, error)
	FindByID(externalID string) (*Destination, error)
	FindBySlug(slug string) (*Destination, error)
	FindByCategory(category string) ([]Destination, error)
	Search(query string) ([]Destination, error)
	Create(dest *Destination) error
	CreateBatch(dests []Destination) error
	Update(dest *Destination) error
	CreateOrUpdateUserDestination(userID uint, slug string, status string) error
	GetUserDestinations(userID uint) ([]UserDestination, error)
	Delete(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll(status string) ([]Destination, error) {
	var dests []Destination
	q := r.db.Order("id ASC")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	err := q.Find(&dests).Error
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
	var err error
	switch category {
	case "hidden-gem":
		err = r.db.Where("status = ? AND rating >= ? AND review_count < ?", "published", 4.5, 2500).Order("rating DESC").Find(&dests).Error
	case "sunset":
		err = r.db.Where("status = ? AND (LOWER(best_time) LIKE ? OR LOWER(best_time) LIKE ?)", "published", "%sore%", "%sunset%").Order("rating DESC").Find(&dests).Error
	case "sunrise":
		err = r.db.Where("status = ? AND (LOWER(best_time) LIKE ? OR LOWER(best_time) LIKE ? OR LOWER(best_time) LIKE ?)", "published", "%sunrise%", "%fajar%", "%dawn%").Order("rating DESC").Find(&dests).Error
	case "camping":
		err = r.db.Where("status = ? AND (LOWER(category) = ? OR LOWER(best_time) LIKE ?)", "published", "camping", "%camping%").Order("rating DESC").Find(&dests).Error
	case "weekend":
		err = r.db.Where("status = ? AND rating >= ? AND review_count >= ?", "published", 4.3, 100).Order("rating DESC").Find(&dests).Error
	case "temple", "candi":
		err = r.db.Where("status = ? AND (LOWER(category) = ? OR LOWER(category) = ? OR LOWER(name) LIKE ? OR LOWER(name) LIKE ? OR LOWER(tagline) LIKE ? OR LOWER(tagline) LIKE ?)", "published", "temple", "candi", "%candi%", "%temple%", "%candi%", "%temple%").Order("rating DESC").Find(&dests).Error
	default:
		err = r.db.Where("status = ? AND category = ?", "published", category).Order("rating DESC").Find(&dests).Error
	}
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
	).Where("status = ?", "published").Order("rating DESC").Find(&dests).Error
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

func (r *GormRepository) CreateOrUpdateUserDestination(userID uint, slug string, status string) error {
	var ud UserDestination
	err := r.db.Where("user_id = ? AND destination_slug = ?", userID, slug).First(&ud).Error
	if err == nil {
		ud.Status = status
		return r.db.Save(&ud).Error
	}
	return r.db.Create(&UserDestination{UserID: userID, DestinationSlug: slug, Status: status}).Error
}

func (r *GormRepository) GetUserDestinations(userID uint) ([]UserDestination, error) {
	var uds []UserDestination
	err := r.db.Where("user_id = ?", userID).Find(&uds).Error
	return uds, err
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Destination{}).Error
}
