package partner

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type JSONArr []interface{}

func (j JSONArr) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *JSONArr) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArr, 0)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan JSONArr: value is not []byte")
	}
	return json.Unmarshal(bytes, j)
}

const (
	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusSuspended = "suspended"
)

type Partner struct {
	gorm.Model
	ExternalID  string  `gorm:"uniqueIndex;not null" json:"id"`
	Name        string  `gorm:"not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	Category    string  `gorm:"index" json:"category"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Image       string  `json:"image"`
	Rating      float64 `json:"rating"`
	Price       string  `json:"price"`
	Distance    string  `json:"distance"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`

	IsSponsored     bool      `gorm:"index" json:"is_sponsored"`
	SponsorTier     int       `json:"sponsor_tier"`
	SponsorStartAt  time.Time `json:"sponsor_start_at"`
	SponsorEndAt    time.Time `json:"sponsor_end_at"`
	TargetDestIDs   JSONArr   `gorm:"type:jsonb" json:"target_dest_ids"`
	ImpressionCount int64     `json:"impression_count"`
	ClickCount      int64     `json:"click_count"`

	SponsorPrice         float64 `json:"sponsor_price"`
	SponsorPriceCurrency string  `gorm:"default:IDR" json:"sponsor_price_currency"`
	SponsorPaymentStatus string  `gorm:"default:pending" json:"sponsor_payment_status"`

	OwnerUserID     *uint      `gorm:"index" json:"owner_user_id"`
	Status          string     `gorm:"size:20;not null;default:approved;index" json:"status"`
	RejectionReason string     `gorm:"type:text" json:"rejection_reason"`
	SubmittedAt     *time.Time `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	ReviewedBy      *uint      `json:"reviewed_by"`
}

type PublicPartnerResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Location    string  `json:"location"`
	Address     string  `json:"address"`
	Image       string  `json:"image"`
	Rating      float64 `json:"rating"`
	Price       string  `json:"price"`
	Distance    string  `json:"distance"`
	Phone       string  `json:"phone"`
	Website     string  `json:"website"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	IsSponsored bool    `json:"is_sponsored"`
	SponsorTier int     `json:"sponsor_tier"`
}

func ToPublicPartnerResponse(partner Partner) PublicPartnerResponse {
	return PublicPartnerResponse{
		ID:          partner.ExternalID,
		Name:        partner.Name,
		Description: partner.Description,
		Category:    partner.Category,
		Location:    partner.Location,
		Address:     partner.Address,
		Image:       partner.Image,
		Rating:      partner.Rating,
		Price:       partner.Price,
		Distance:    partner.Distance,
		Phone:       partner.Phone,
		Website:     partner.Website,
		Latitude:    partner.Latitude,
		Longitude:   partner.Longitude,
		IsSponsored: partner.IsSponsored,
		SponsorTier: partner.SponsorTier,
	}
}

func ToPublicPartnerResponses(partners []Partner) []PublicPartnerResponse {
	responses := make([]PublicPartnerResponse, 0, len(partners))
	for _, partner := range partners {
		responses = append(responses, ToPublicPartnerResponse(partner))
	}
	return responses
}

func (p *Partner) IsSponsorshipActive(now time.Time) bool {
	if !p.IsSponsored {
		return false
	}
	if !p.SponsorStartAt.IsZero() && now.Before(p.SponsorStartAt) {
		return false
	}
	if !p.SponsorEndAt.IsZero() && now.After(p.SponsorEndAt) {
		return false
	}
	return true
}
