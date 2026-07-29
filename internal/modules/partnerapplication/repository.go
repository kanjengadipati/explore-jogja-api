package partnerapplication

import "gorm.io/gorm"

type Repository interface {
	Create(app *PartnerApplication) error
	FindByExternalID(externalID string) (*PartnerApplication, error)
	FindByStatus(status string) ([]PartnerApplication, error)
	FindByApplicant(userID uint) ([]PartnerApplication, error)
	Update(app *PartnerApplication) error
}

type GormRepository struct{ db *gorm.DB }

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) Create(app *PartnerApplication) error { return r.db.Create(app).Error }

func (r *GormRepository) FindByExternalID(externalID string) (*PartnerApplication, error) {
	var app PartnerApplication
	err := r.db.Where("external_id = ?", externalID).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *GormRepository) FindByStatus(status string) ([]PartnerApplication, error) {
	var apps []PartnerApplication
	err := r.db.Where("status = ?", status).Order("id ASC").Find(&apps).Error
	return apps, err
}

func (r *GormRepository) FindByApplicant(userID uint) ([]PartnerApplication, error) {
	var apps []PartnerApplication
	err := r.db.Where("applicant_user_id = ?", userID).Order("id DESC").Find(&apps).Error
	return apps, err
}

func (r *GormRepository) Update(app *PartnerApplication) error { return r.db.Save(app).Error }
