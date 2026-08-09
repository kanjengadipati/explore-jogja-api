package bonus

const timeLayout = "2006-01-02T15:04:05Z07:00"

type BonusResponse struct {
	ID             uint    `json:"id"`
	SalesUserID    uint    `json:"sales_user_id"`
	SalesUserName  string  `json:"sales_user_name,omitempty"`
	SalesUserEmail string  `json:"sales_user_email,omitempty"`
	Type           string  `json:"type"`
	PartnerUserID *uint   `json:"partner_user_id,omitempty"`
	Period         *string `json:"period,omitempty"`
	Metric         string  `json:"metric,omitempty"`
	Tier           *int    `json:"tier,omitempty"`
	Amount         float64 `json:"amount"`
	Status         string  `json:"status"`
	CreatedAt      string  `json:"created_at"`
}

func toBonusResponse(b *SalesBonus) BonusResponse {
	return BonusResponse{
		ID:           b.ID,
		SalesUserID:  b.SalesUserID,
		Type:         string(b.Type),
		PartnerUserID: b.PartnerUserID,
		Period:       b.Period,
		Metric:       string(b.Metric),
		Tier:         b.Tier,
		Amount:       b.Amount,
		Status:       string(b.Status),
		CreatedAt:    b.CreatedAt.Format(timeLayout),
	}
}

type BonusRuleResponse struct {
	ID             uint    `json:"id"`
	Type           string  `json:"type"`
	Metric         string  `json:"metric"`
	Tier           *int    `json:"tier,omitempty"`
	Threshold      *int    `json:"threshold,omitempty"`
	Amount         float64 `json:"amount"`
	IsActive       bool    `json:"is_active"`
	EffectiveFrom  *string `json:"effective_from,omitempty"`
	EffectiveUntil *string `json:"effective_until,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func toBonusRuleResponse(r *BonusRule) BonusRuleResponse {
	res := BonusRuleResponse{
		ID:        r.ID,
		Type:      string(r.Type),
		Metric:    string(r.Metric),
		Tier:      r.Tier,
		Threshold: r.Threshold,
		Amount:    r.Amount,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt.Format(timeLayout),
		UpdatedAt: r.UpdatedAt.Format(timeLayout),
	}
	if r.EffectiveFrom != nil {
		s := r.EffectiveFrom.Format("2006-01-02")
		res.EffectiveFrom = &s
	}
	if r.EffectiveUntil != nil {
		s := r.EffectiveUntil.Format("2006-01-02")
		res.EffectiveUntil = &s
	}
	return res
}

// CreateBonusRuleRequest is used for both create and update (full update).
type CreateBonusRuleRequest struct {
	Type           BonusType   `json:"type" binding:"required,oneof=onboarding milestone"`
	Metric         BonusMetric `json:"metric"`
	Tier           *int        `json:"tier"`
	Threshold      *int        `json:"threshold"`
	Amount         float64     `json:"amount" binding:"required,gt=0"`
	IsActive       *bool       `json:"is_active"`
	EffectiveFrom  *string     `json:"effective_from"`
	EffectiveUntil *string     `json:"effective_until"`
}

// UpdateBonusStatusRequest is the admin payout action: mark a pending bonus as
// paid (money actually transferred) or voided (payout cancelled).
type UpdateBonusStatusRequest struct {
	Status Status `json:"status" binding:"required,oneof=paid voided"`
}
