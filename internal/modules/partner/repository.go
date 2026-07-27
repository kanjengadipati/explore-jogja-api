package partner

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Partner, error)
	FindAllApproved() ([]Partner, error)
	FindByID(externalID string) (*Partner, error)
	FindByIDAny(externalID string) (*Partner, error)
	FindByIDAndOwner(externalID string, ownerUserID uint) (*Partner, error)
	FindByOwnerID(ownerUserID uint) ([]Partner, error)
	FindByStatus(status string) ([]Partner, error)
	Search(query string) ([]Partner, error)
	Create(partner *Partner) error
	Update(partner *Partner) error
	Delete(externalID string) error
	DeleteByIDAndOwner(externalID string, ownerUserID uint) error
	FindSponsored(destinationID, category string) ([]Partner, error)
	IncrementImpression(externalID string) error
	IncrementClick(externalID string) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

// --- Public (approved only) ---

func (r *GormRepository) FindAll() ([]Partner, error) {
	var partners []Partner
	err := r.db.Where("status = ?", StatusApproved).Order("id ASC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) FindAllApproved() ([]Partner, error) {
	var partners []Partner
	err := r.db.Where("status = ?", StatusApproved).Order("id ASC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) FindByID(externalID string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("external_id = ? AND status = ?", externalID, StatusApproved).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

// --- Internal / Admin (any status) ---

func (r *GormRepository) FindByIDAny(externalID string) (*Partner, error) {
	var partner Partner
	err := r.db.Where("external_id = ?", externalID).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *GormRepository) FindByStatus(status string) ([]Partner, error) {
	var partners []Partner
	err := r.db.Where("status = ?", status).Order("submitted_at ASC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) Search(query string) ([]Partner, error) {
	var partners []Partner
	like := "%" + query + "%"
	err := r.db.Where("status = ?", StatusApproved).
		Where(
			r.db.Where("name ILIKE ?", like).
				Or("description ILIKE ?", like).
				Or("location ILIKE ?", like).
				Or("category ILIKE ?", like),
		).Order("rating DESC").Find(&partners).Error
	return partners, err
}

// --- Owner-scoped ---

func (r *GormRepository) FindByIDAndOwner(externalID string, ownerUserID uint) (*Partner, error) {
	var partner Partner
	err := r.db.Where("external_id = ? AND owner_user_id = ?", externalID, ownerUserID).First(&partner).Error
	if err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *GormRepository) FindByOwnerID(ownerUserID uint) ([]Partner, error) {
	var partners []Partner
	err := r.db.Where("owner_user_id = ?", ownerUserID).Order("id ASC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) DeleteByIDAndOwner(externalID string, ownerUserID uint) error {
	return r.db.Where("external_id = ? AND owner_user_id = ?", externalID, ownerUserID).Delete(&Partner{}).Error
}

// --- Write ---

func (r *GormRepository) Create(partner *Partner) error {
	return r.db.Create(partner).Error
}

func (r *GormRepository) Update(partner *Partner) error {
	return r.db.Save(partner).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Partner{}).Error
}

// --- Sponsored / Tracking ---

func (r *GormRepository) FindSponsored(destinationID, category string) ([]Partner, error) {
	now := time.Now()
	zero := time.Time{}
	q := r.db.Where("is_sponsored = ? AND status = ?", true, StatusApproved).
		Where("(sponsor_start_at IS NULL OR sponsor_start_at = ? OR sponsor_start_at <= ?)", zero, now).
		Where("(sponsor_end_at IS NULL OR sponsor_end_at = ? OR sponsor_end_at >= ?)", zero, now)

	if category != "" {
		q = q.Where("category = ?", category)
	}
	if destinationID != "" {
		q = q.Where("target_dest_ids @> ? OR target_dest_ids = '[]' OR target_dest_ids IS NULL", `["`+destinationID+`"]`)
	}

	var partners []Partner
	err := q.Order("sponsor_tier ASC, rating DESC").Find(&partners).Error
	return partners, err
}

func (r *GormRepository) IncrementImpression(externalID string) error {
	return r.db.Model(&Partner{}).
		Where("external_id = ?", externalID).
		UpdateColumn("impression_count", gorm.Expr("impression_count + 1")).Error
}

func (r *GormRepository) IncrementClick(externalID string) error {
	return r.db.Model(&Partner{}).
		Where("external_id = ?", externalID).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error
}
