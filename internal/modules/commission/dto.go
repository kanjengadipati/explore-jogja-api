package commission

type CommissionResponse struct {
	ID                uint    `json:"id"`
	PartnerUserID     uint    `json:"partner_user_id"`
	OrderID           string  `json:"order_id"`
	SubjectType       string  `json:"subject_type"`
	GrossAmount       float64 `json:"gross_amount"`
	CommissionRate    float64 `json:"commission_rate"`
	CommissionAmount  float64 `json:"commission_amount"`
	Status            string  `json:"status"`
	CreatedAt         string  `json:"created_at"`
}

func toCommissionResponse(c *SalesCommission) CommissionResponse {
	return CommissionResponse{
		ID:               c.ID,
		PartnerUserID:    c.PartnerUserID,
		OrderID:          c.OrderID,
		SubjectType:      c.SubjectType,
		GrossAmount:      c.GrossAmount,
		CommissionRate:   c.CommissionRate,
		CommissionAmount: c.CommissionAmount,
		Status:           string(c.Status),
		CreatedAt:        c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// SalesPerformanceItem is one row of the superadmin sales-performance report.
type SalesPerformanceItem struct {
	SalesUserID           uint    `json:"sales_user_id"`
	SalesName             string  `json:"sales_name"`
	SalesEmail            string  `json:"sales_email"`
	ReferralCode          string  `json:"referral_code"`
	TotalPartners         int64   `json:"total_partners"`
	TotalTransactions     int64   `json:"total_transactions"`
	TotalVolume           float64 `json:"total_volume"`
	VolumeFromSubscription float64 `json:"volume_from_subscription"`
	VolumeFromAdCampaign   float64 `json:"volume_from_ad_campaign"`
	PendingCommission     float64 `json:"pending_commission"`
	PaidCommission        float64 `json:"paid_commission"`
	TotalCommission       float64 `json:"total_commission_earned"`
}

type UpdateCommissionRateRequest struct {
	// Rate as a fraction, e.g. 0.20 for 20%.
	Rate float64 `json:"rate" binding:"required,gt=0,lt=1"`
}
