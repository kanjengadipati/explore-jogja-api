package commission

import (
	"strconv"

	"pleco-api/internal/modules/config"
	"pleco-api/internal/modules/user"
)

// ConfigRateKey is where the admin-editable commission rate lives in the
// generic SiteConfig key/value store (category "sales"). Value is a plain
// decimal string, e.g. "0.20" for 20%.
const ConfigRateKey = "sales_commission_rate"

type Service struct {
	repo      Repository
	configSvc *config.Service
	userSvc   *user.Service
}

func NewService(repo Repository, configSvc *config.Service, userSvc *user.Service) *Service {
	return &Service{repo: repo, configSvc: configSvc, userSvc: userSvc}
}

// GetCommissionRate returns the current admin-configured rate, falling back
// to DefaultRate if it was never set or fails to parse.
func (s *Service) GetCommissionRate() float64 {
	cfg, err := s.configSvc.GetByKey(ConfigRateKey)
	if err != nil || cfg.Value == "" {
		return DefaultRate
	}
	rate, err := strconv.ParseFloat(cfg.Value, 64)
	if err != nil || rate <= 0 || rate >= 1 {
		return DefaultRate
	}
	return rate
}

// SetCommissionRate updates the admin-configured rate (fraction, e.g. 0.20).
// Only affects commissions recorded after this call — past rows keep their
// own snapshot rate.
func (s *Service) SetCommissionRate(rate float64) error {
	_, err := s.configSvc.Update(config.UpdateSiteConfigRequest{
		Key:      ConfigRateKey,
		Value:    strconv.FormatFloat(rate, 'f', 4, 64),
		Category: "sales",
	})
	return err
}

// RecordFromTransaction creates a commission row for a paid transaction, if
// (and only if) the payer was referred by a sales user. Safe to call more
// than once for the same transaction — a no-op if a commission already
// exists for this PaymentTransactionID.
func (s *Service) RecordFromTransaction(payerUserID, paymentTransactionID uint, orderID, subjectType string, grossAmount float64) error {
	if _, err := s.repo.GetByPaymentTransactionID(paymentTransactionID); err == nil {
		return nil // already recorded
	}

	payer, err := s.userSvc.UserRepo.FindByID(payerUserID)
	if err != nil || payer.ReferredBySalesID == nil {
		return nil // payer not found, or not referred by anyone — no commission owed
	}

	rate := s.GetCommissionRate()
	c := &SalesCommission{
		SalesUserID:          *payer.ReferredBySalesID,
		PartnerUserID:        payerUserID,
		PaymentTransactionID: paymentTransactionID,
		OrderID:              orderID,
		SubjectType:          subjectType,
		GrossAmount:          grossAmount,
		CommissionRate:       rate,
		CommissionAmount:     grossAmount * rate,
		Status:               StatusPending,
	}
	return s.repo.Create(c)
}

// VoidFromTransaction voids a previously recorded commission. Not wired to
// any trigger yet — jogjagem's payment module doesn't have a refund flow at
// the moment — but ready for when one exists.
func (s *Service) VoidFromTransaction(paymentTransactionID uint) error {
	return s.repo.VoidByPaymentTransactionID(paymentTransactionID)
}

// ListMyCommissions returns a sales user's own commission history.
func (s *Service) ListMyCommissions(salesUserID uint, limit, offset int) ([]CommissionResponse, int64, error) {
	items, total, err := s.repo.ListBySales(salesUserID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	res := make([]CommissionResponse, len(items))
	for i := range items {
		res[i] = toCommissionResponse(&items[i])
	}
	return res, total, nil
}

// GetSalesPerformanceReport builds the superadmin report: one row per sales
// user with their referred-partner count, transaction volume (broken down by
// subscription vs ad campaign), and commission earned (pending vs. paid).
func (s *Service) GetSalesPerformanceReport() ([]SalesPerformanceItem, error) {
	salesUsers, _, err := s.userSvc.GetAllUsers(1, 1000, "", "sales")
	if err != nil {
		return nil, err
	}

	report := make([]SalesPerformanceItem, 0, len(salesUsers))
	for _, su := range salesUsers {
		pending, paid, err := s.repo.SumBySales(su.ID)
		if err != nil {
			return nil, err
		}
		txCount, err := s.repo.CountBySales(su.ID)
		if err != nil {
			return nil, err
		}
		partnerCount, err := s.repo.CountDistinctPartnersBySales(su.ID)
		if err != nil {
			return nil, err
		}
		totalVol, subVol, adVol, err := s.repo.SumVolumeBySales(su.ID)
		if err != nil {
			return nil, err
		}

		code := ""
		if su.ReferralCode != nil {
			code = *su.ReferralCode
		}

		report = append(report, SalesPerformanceItem{
			SalesUserID:            su.ID,
			SalesName:              su.Name,
			SalesEmail:             su.Email,
			ReferralCode:           code,
			TotalPartners:          partnerCount,
			TotalTransactions:      txCount,
			TotalVolume:            totalVol,
			VolumeFromSubscription: subVol,
			VolumeFromAdCampaign:   adVol,
			PendingCommission:      pending,
			PaidCommission:         paid,
			TotalCommission:        pending + paid,
		})
	}
	return report, nil
}
