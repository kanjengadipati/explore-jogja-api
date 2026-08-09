package commission

import (
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusVoided  Status = "voided" // reserved for a future refund flow
)

// DefaultRate is used only if no admin-configured rate exists yet (see
// config.SiteConfig key "sales_commission_rate").
const DefaultRate = 0.20

// SalesCommission is created once per paid PaymentTransaction whose payer
// (CreatedByUserID) was referred by a sales user. SalesUserID is a snapshot:
// if the partner's referral is ever reassigned later, past commission rows
// keep pointing to the sales person who actually earned them.
type SalesCommission struct {
	gorm.Model
	SalesUserID          uint      `gorm:"not null;index" json:"sales_user_id"`
	PartnerUserID         uint      `gorm:"not null;index" json:"partner_user_id"`
	PaymentTransactionID uint      `gorm:"not null;uniqueIndex" json:"payment_transaction_id"`
	OrderID              string    `gorm:"size:100;not null" json:"order_id"`
	SubjectType           string    `gorm:"size:50;not null;index" json:"subject_type"` // subscription | ad_campaign
	GrossAmount          float64   `gorm:"not null" json:"gross_amount"`
	CommissionRate       float64   `gorm:"not null" json:"commission_rate"`
	CommissionAmount     float64   `gorm:"not null" json:"commission_amount"`
	Status               Status    `gorm:"size:20;not null;default:pending;index" json:"status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (SalesCommission) TableName() string {
	return "sales_commissions"
}
