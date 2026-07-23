package imagereport

import "gorm.io/gorm"

type Repository interface {
	Create(report *ImageReport) error
	FindAll() ([]ImageReport, error)
	FindByDestinationID(destID string) ([]ImageReport, error)
	FindByStatus(status string) ([]ImageReport, error)
	UpdateStatus(id uint, status string) error
	CountByStatus(status string) (int64, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(report *ImageReport) error {
	return r.db.Create(report).Error
}

func (r *gormRepository) FindAll() ([]ImageReport, error) {
	var reports []ImageReport
	err := r.db.
		Table("image_reports").
		Select("image_reports.*, COALESCE(destinations.name, '') as destination_name, COALESCE(destinations.location, '') as destination_location").
		Joins("LEFT JOIN destinations ON destinations.external_id = image_reports.destination_id AND destinations.deleted_at IS NULL").
		Where("image_reports.deleted_at IS NULL").
		Order("image_reports.created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *gormRepository) FindByDestinationID(destID string) ([]ImageReport, error) {
	var reports []ImageReport
	err := r.db.
		Table("image_reports").
		Select("image_reports.*, COALESCE(destinations.name, '') as destination_name, COALESCE(destinations.location, '') as destination_location").
		Joins("LEFT JOIN destinations ON destinations.external_id = image_reports.destination_id AND destinations.deleted_at IS NULL").
		Where("image_reports.destination_id = ? AND image_reports.deleted_at IS NULL", destID).
		Order("image_reports.created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *gormRepository) FindByStatus(status string) ([]ImageReport, error) {
	var reports []ImageReport
	err := r.db.
		Table("image_reports").
		Select("image_reports.*, COALESCE(destinations.name, '') as destination_name, COALESCE(destinations.location, '') as destination_location").
		Joins("LEFT JOIN destinations ON destinations.external_id = image_reports.destination_id AND destinations.deleted_at IS NULL").
		Where("image_reports.status = ? AND image_reports.deleted_at IS NULL", status).
		Order("image_reports.created_at DESC").
		Find(&reports).Error
	return reports, err
}

func (r *gormRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&ImageReport{}).Where("id = ?", id).Update("status", status).Error
}

func (r *gormRepository) CountByStatus(status string) (int64, error) {
	var count int64
	err := r.db.Model(&ImageReport{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
