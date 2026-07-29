package partnerapplication

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

type PartnerApplication struct {
	gorm.Model
	ExternalID      string `gorm:"uniqueIndex;not null" json:"id"`
	ApplicantUserID uint   `gorm:"index;not null" json:"applicant_user_id"`

	BusinessName string `gorm:"not null" json:"business_name"`
	Category     string `gorm:"not null" json:"category"`
	Location     string `json:"location"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`

	Status          string `gorm:"default:pending;index" json:"status"`
	RejectionReason string `gorm:"type:text" json:"rejection_reason"`

	ConvertedPartnerExternalID *string `json:"converted_partner_external_id,omitempty"`

	SubmittedAt *time.Time `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	ReviewedBy  *uint      `json:"reviewed_by"`
}
