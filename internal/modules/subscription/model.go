package subscription

import (
	"time"

	"gorm.io/gorm"
)

const (
	PlanFree         = "free"
	PlanPro          = "pro"
	PlanBusinessPlus = "business_plus"
	PlanEnterprise   = "enterprise"

	StatusActive   = "active"
	StatusPastDue  = "past_due"
	StatusCanceled = "canceled"
)

type Subscription struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	ExternalID       string         `gorm:"uniqueIndex;not null" json:"external_id"`
	BusinessID       uint           `gorm:"not null" json:"business_id"`
	Plan             string         `gorm:"not null;default:free" json:"plan"`
	Status           string         `gorm:"not null;default:active" json:"status"`
	CurrentPeriodEnd *time.Time     `json:"current_period_end"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (s *Subscription) CanCreateAdCampaign() bool {
	// A Free-plan business is blocked from creating ad campaigns
	return s.Plan != PlanFree && s.Status == StatusActive
}
