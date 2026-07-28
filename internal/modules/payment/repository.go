package payment

import "gorm.io/gorm"

type Repository interface {
	Create(tx *PaymentTransaction) error
	FindByOrderID(orderID string) (*PaymentTransaction, error)
	FindAll() ([]PaymentTransaction, error)
	FindBySubject(subjectType, subjectExternalID string) ([]PaymentTransaction, error)
	FindPendingExpired() ([]PaymentTransaction, error)
	Update(tx *PaymentTransaction) error
}

type GormRepository struct {
	db *gorm.DB
}

var _ Repository = (*GormRepository)(nil)

func NewRepository(db *gorm.DB) Repository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(tx *PaymentTransaction) error {
	return r.db.Create(tx).Error
}

func (r *GormRepository) FindByOrderID(orderID string) (*PaymentTransaction, error) {
	var tx PaymentTransaction
	err := r.db.Where("order_id = ?", orderID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *GormRepository) FindAll() ([]PaymentTransaction, error) {
	var txs []PaymentTransaction
	err := r.db.Order("id DESC").Find(&txs).Error
	return txs, err
}

func (r *GormRepository) FindBySubject(subjectType, subjectExternalID string) ([]PaymentTransaction, error) {
	var txs []PaymentTransaction
	err := r.db.Where("subject_type = ? AND subject_external_id = ?", subjectType, subjectExternalID).
		Order("id DESC").Find(&txs).Error
	return txs, err
}

func (r *GormRepository) FindPendingExpired() ([]PaymentTransaction, error) {
	var txs []PaymentTransaction
	err := r.db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < NOW()", StatusPending).
		Find(&txs).Error
	return txs, err
}

func (r *GormRepository) Update(tx *PaymentTransaction) error {
	return r.db.Save(tx).Error
}
