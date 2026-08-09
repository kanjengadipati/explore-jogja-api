package commission

import "gorm.io/gorm"

type Repository interface {
	Create(c *SalesCommission) error
	GetByPaymentTransactionID(paymentTransactionID uint) (*SalesCommission, error)
	FindByID(id uint) (*SalesCommission, error)
	UpdateStatus(id uint, status Status) error
	VoidByPaymentTransactionID(paymentTransactionID uint) error
	ListBySales(salesUserID uint, limit, offset int) ([]SalesCommission, int64, error)
	SumBySales(salesUserID uint) (pending float64, paid float64, err error)
	CountBySales(salesUserID uint) (int64, error)
	CountDistinctPartnersBySales(salesUserID uint) (int64, error)
	SumVolumeBySales(salesUserID uint) (total float64, err error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(c *SalesCommission) error {
	return r.db.Create(c).Error
}

func (r *repository) GetByPaymentTransactionID(paymentTransactionID uint) (*SalesCommission, error) {
	var c SalesCommission
	err := r.db.Where("payment_transaction_id = ?", paymentTransactionID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repository) FindByID(id uint) (*SalesCommission, error) {
	var c SalesCommission
	err := r.db.First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repository) UpdateStatus(id uint, status Status) error {
	return r.db.Model(&SalesCommission{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) VoidByPaymentTransactionID(paymentTransactionID uint) error {
	return r.db.Model(&SalesCommission{}).
		Where("payment_transaction_id = ? AND status != ?", paymentTransactionID, StatusVoided).
		Update("status", StatusVoided).Error
}

func (r *repository) ListBySales(salesUserID uint, limit, offset int) ([]SalesCommission, int64, error) {
	var total int64
	r.db.Model(&SalesCommission{}).Where("sales_user_id = ?", salesUserID).Count(&total)

	var items []SalesCommission
	err := r.db.Where("sales_user_id = ?", salesUserID).
		Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *repository) SumBySales(salesUserID uint) (float64, float64, error) {
	var pending, paid float64
	if err := r.db.Model(&SalesCommission{}).
		Where("sales_user_id = ? AND status = ?", salesUserID, StatusPending).
		Select("COALESCE(SUM(commission_amount),0)").Scan(&pending).Error; err != nil {
		return 0, 0, err
	}
	if err := r.db.Model(&SalesCommission{}).
		Where("sales_user_id = ? AND status = ?", salesUserID, StatusPaid).
		Select("COALESCE(SUM(commission_amount),0)").Scan(&paid).Error; err != nil {
		return 0, 0, err
	}
	return pending, paid, nil
}

func (r *repository) CountBySales(salesUserID uint) (int64, error) {
	var count int64
	err := r.db.Model(&SalesCommission{}).
		Where("sales_user_id = ? AND status != ?", salesUserID, StatusVoided).
		Count(&count).Error
	return count, err
}

func (r *repository) CountDistinctPartnersBySales(salesUserID uint) (int64, error) {
	var count int64
	err := r.db.Model(&SalesCommission{}).
		Where("sales_user_id = ?", salesUserID).
		Distinct("partner_user_id").Count(&count).Error
	return count, err
}

func (r *repository) SumVolumeBySales(salesUserID uint) (float64, error) {
	var total float64
	err := r.db.Model(&SalesCommission{}).
		Where("sales_user_id = ? AND status != ?", salesUserID, StatusVoided).
		Select("COALESCE(SUM(gross_amount),0)").Scan(&total).Error
	return total, err
}
