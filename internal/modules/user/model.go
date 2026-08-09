package user

import (
	"time"

	roleModule "pleco-api/internal/modules/role"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name               string
	Email              string          `gorm:"unique" json:"email"`
	PhoneNumber        string          `gorm:"unique" json:"phone_number,omitempty"`
	Password           string          `json:"-"`
	Role               string          `json:"role"` // user / admin
	RoleID             uint            `json:"role_id"`
	RoleDetails        roleModule.Role `gorm:"foreignKey:RoleID" json:"role_details,omitempty"`
	IsVerified         bool            `json:"is_verified"`
	PhoneVerified      bool            `json:"phone_verified"`
	EmailVerified      bool            `json:"email_verified"`
	AvatarURL          string          `gorm:"type:text" json:"avatar_url,omitempty"`
	CoverImageURL      string          `gorm:"type:text" json:"cover_image_url,omitempty"`
	PasswordUpdatedAt  time.Time
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	LastPasswordChange *time.Time `json:"last_password_change_at,omitempty"`
	AccessTokenVersion uint       `gorm:"default:0" json:"-"`
	// ReferralCode is only meaningful for Role=="sales" — their referral link/code.
	ReferralCode *string `gorm:"size:20;uniqueIndex" json:"referral_code,omitempty"`
	// ReferredBySalesID is set once, when a partner signs up (or creates their
	// first business) through a sales referral code. It's a snapshot — commission
	// rows keep their own SalesUserID snapshot too, so this can be reassigned
	// later without rewriting past commissions.
	ReferredBySalesID *uint `gorm:"index" json:"referred_by_sales_id,omitempty"`
	// ReferredAt is when ReferredBySalesID was set — used to compute which
	// commission tier a payment falls into (see commission.Service). Existing
	// partners backfilled from created_at by migration 000092.
	ReferredAt *time.Time `json:"referred_at,omitempty"`
}
