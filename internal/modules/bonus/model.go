package bonus

import (
	"time"

	"gorm.io/gorm"
)

type BonusType string

const (
	BonusTypeOnboarding BonusType = "onboarding"
	BonusTypeMilestone  BonusType = "milestone"
)

type BonusMetric string

const (
	MetricTenant      BonusMetric = "tenant"
	MetricTransaction BonusMetric = "transaction"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusPaid    Status = "paid"
	StatusVoided  Status = "voided"
)

// SalesBonus is created per bonus event:
//   - onboarding: one row per (sales_user_id, partner_user_id), unique partial index.
//   - milestone: one row per (sales_user_id, period, metric, tier), unique partial index.
//
// Status lifecycle: pending → paid (manual payout), or pending → voided (e.g. the
// partner's first transaction was refunded). Rows are never hard-deleted.
type SalesBonus struct {
	gorm.Model
	SalesUserID  uint        `gorm:"not null;index" json:"sales_user_id"`
	Type         BonusType   `gorm:"size:20;not null;index" json:"type"`
	PartnerUserID *uint      `gorm:"index" json:"partner_user_id,omitempty"`
	Period       *string     `gorm:"size:7" json:"period,omitempty"`
	Metric       BonusMetric `gorm:"size:20" json:"metric,omitempty"`
	Tier         *int        `json:"tier,omitempty"`
	Amount       float64     `gorm:"not null" json:"amount"`
	Status       Status      `gorm:"size:20;not null;default:pending;index" json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func (SalesBonus) TableName() string {
	return "sales_bonuses"
}

type SalesUserInfo struct {
	Name  string
	Email string
}

// BonusRule is the admin-editable configuration. An onboarding rule carries a
// flat amount (tier/threshold ignored); milestone rules define tiers with a
// threshold and an amount. Only active rules within their effective window apply.
type BonusRule struct {
	gorm.Model
	Type           BonusType   `gorm:"size:20;not null;index" json:"type"`
	Metric         BonusMetric `gorm:"size:20;not null;default:tenant" json:"metric"`
	Tier           *int        `json:"tier,omitempty"`
	Threshold      *int        `json:"threshold,omitempty"`
	Amount         float64     `gorm:"not null" json:"amount"`
	IsActive       bool        `gorm:"not null;default:true" json:"is_active"`
	EffectiveFrom  *time.Time  `json:"effective_from,omitempty"`
	EffectiveUntil *time.Time  `json:"effective_until,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (BonusRule) TableName() string {
	return "bonus_rules"
}
