package payment

import (
	"time"

	"gorm.io/gorm"
)

const (
	SubjectAdCampaign         = "ad_campaign"
	SubjectPartnerSponsorship = "partner_sponsorship"
	SubjectSubscription       = "subscription"

	StatusPending  = "pending"
	StatusPaid     = "paid"
	StatusExpired  = "expired"
	StatusFailed   = "failed"
	StatusRefunded = "refunded"
)

type PaymentTransaction struct {
	gorm.Model
	OrderID           string `gorm:"uniqueIndex;not null" json:"order_id"`
	SubjectType       string `gorm:"index;not null" json:"subject_type"`
	SubjectExternalID string `gorm:"index;not null" json:"subject_external_id"`

	Amount   float64 `json:"amount"`
	Currency string  `gorm:"default:IDR" json:"currency"`

	Status        string `gorm:"default:pending;index" json:"status"`
	MidtransToken string `json:"midtrans_token"`
	PaymentType   string `json:"payment_type"`
	TransactionID string `json:"transaction_id"`
	VANumber      string `json:"va_number,omitempty"`
	FraudStatus   string `json:"fraud_status,omitempty"`

	RawNotification string     `gorm:"type:text" json:"-"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`

	CreatedByUserID *uint `json:"created_by_user_id,omitempty"`
}
