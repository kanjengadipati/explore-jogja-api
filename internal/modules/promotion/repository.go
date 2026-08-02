package promotion

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindAll() ([]Promotion, error)
	FindByID(externalID string) (*Promotion, error)
	FindByIDAny(externalID string) (*Promotion, error)
	FindByStatus(status string) ([]Promotion, error)
	Search(query string) ([]Promotion, error)
	Create(promotion *Promotion) error
	Update(promotion *Promotion) error
	Delete(externalID string) error
	FindByPartnerID(partnerExternalID string) ([]Promotion, error)
	FindByIDAndPartner(externalID string, partnerExternalID string) (*Promotion, error)
	FindByBusinessID(businessExternalID string) ([]Promotion, error)
	FindByIDAndBusiness(externalID string, businessExternalID string) (*Promotion, error)
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindAll() ([]Promotion, error) {
	var promotions []Promotion
	err := r.db.Where(
		"status = 'approved' AND (business_external_id IS NULL OR business_external_id IN (SELECT external_id FROM businesses WHERE status = 'approved'))",
	).Order("id ASC").Find(&promotions).Error
	return promotions, err
}

func (r *GormRepository) FindByID(externalID string) (*Promotion, error) {
	var promotion Promotion
	err := r.db.Where(
		"external_id = ? AND (business_external_id IS NULL OR business_external_id IN (SELECT external_id FROM businesses WHERE status = 'approved'))",
		externalID,
	).First(&promotion).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (r *GormRepository) FindByIDAny(externalID string) (*Promotion, error) {
	var promotion Promotion
	err := r.db.Where("external_id = ?", externalID).First(&promotion).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (r *GormRepository) Search(query string) ([]Promotion, error) {
	var promotions []Promotion
	like := "%" + query + "%"
	err := r.db.Where(
		"business_external_id IS NULL OR business_external_id IN (SELECT external_id FROM businesses WHERE status = 'approved')",
	).Where(
		r.db.Where("title ILIKE ?", like).
			Or("description ILIKE ?", like).
			Or("category ILIKE ?", like).
			Or("code ILIKE ?", like),
	).Order("start_date DESC").Find(&promotions).Error
	return promotions, err
}

func (r *GormRepository) Create(promotion *Promotion) error {
	return r.db.Create(promotion).Error
}

func (r *GormRepository) Update(promotion *Promotion) error {
	return r.db.Save(promotion).Error
}

func (r *GormRepository) Delete(externalID string) error {
	return r.db.Where("external_id = ?", externalID).Delete(&Promotion{}).Error
}

func (r *GormRepository) FindByPartnerID(partnerExternalID string) ([]Promotion, error) {
	var promotions []Promotion
	err := r.db.Where("legacy_partner_external_id = ?", partnerExternalID).Order("id ASC").Find(&promotions).Error
	return promotions, err
}

func (r *GormRepository) FindByIDAndPartner(externalID string, partnerExternalID string) (*Promotion, error) {
	var promotion Promotion
	err := r.db.Where("external_id = ? AND legacy_partner_external_id = ?", externalID, partnerExternalID).First(&promotion).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (r *GormRepository) FindByBusinessID(businessExternalID string) ([]Promotion, error) {
	var promotions []Promotion
	err := r.db.Where("business_external_id = ?", businessExternalID).Order("id ASC").Find(&promotions).Error
	return promotions, err
}

func (r *GormRepository) FindByIDAndBusiness(externalID string, businessExternalID string) (*Promotion, error) {
	var promotion Promotion
	err := r.db.Where("external_id = ? AND business_external_id = ?", externalID, businessExternalID).First(&promotion).Error
	if err != nil {
		return nil, err
	}
	return &promotion, nil
}

func (r *GormRepository) FindByStatus(status string) ([]Promotion, error) {
	var promotions []Promotion
	err := r.db.Where("status = ?", status).Order("id ASC").Find(&promotions).Error
	return promotions, err
}
