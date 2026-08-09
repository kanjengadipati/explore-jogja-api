package commission

import (
	"strconv"
	"time"

	"pleco-api/internal/modules/config"
	"pleco-api/internal/modules/user"
)

// Config keys untuk 2 tier — keduanya bisa diubah admin lewat
// GET/PUT /admin/sales-commission-rate.
const (
	ConfigRateTier1Key = "sales_commission_rate_tier1" // TierThresholdMonths pertama
	ConfigRateTier2Key = "sales_commission_rate_tier2" // setelah itu, recurring
)

// TierThresholdMonths — lama tier 1 berlaku, dihitung dari User.ReferredAt.
// Setelah ini, tier 2 berlaku seumur hidup partner.
const TierThresholdMonths = 12

type Service struct {
	repo      Repository
	configSvc *config.Service
	userSvc   *user.Service
}

func NewService(repo Repository, configSvc *config.Service, userSvc *user.Service) *Service {
	return &Service{repo: repo, configSvc: configSvc, userSvc: userSvc}
}

// GetCommissionRates mengembalikan rate tier 1 & tier 2 yang aktif saat ini,
// fallback ke default kalau belum pernah di-set admin / gagal parse.
func (s *Service) GetCommissionRates() (tier1, tier2 float64) {
	tier1 = s.readRate(ConfigRateTier1Key, DefaultRateTier1)
	tier2 = s.readRate(ConfigRateTier2Key, DefaultRateTier2)
	return tier1, tier2
}

func (s *Service) readRate(key string, fallback float64) float64 {
	cfg, err := s.configSvc.GetByKey(key)
	if err != nil || cfg.Value == "" {
		return fallback
	}
	rate, err := strconv.ParseFloat(cfg.Value, 64)
	if err != nil || rate <= 0 || rate >= 1 {
		return fallback
	}
	return rate
}

// SetCommissionRates meng-update kedua rate tier (fraction, mis. 0.20).
// Hanya berlaku untuk komisi yang direcord SETELAH pemanggilan ini — baris
// lama tetap simpan snapshot rate-nya sendiri.
func (s *Service) SetCommissionRates(tier1, tier2 float64) error {
	if _, err := s.configSvc.Update(config.UpdateSiteConfigRequest{
		Key: ConfigRateTier1Key, Value: strconv.FormatFloat(tier1, 'f', 4, 64), Category: "sales",
	}); err != nil {
		return err
	}
	_, err := s.configSvc.Update(config.UpdateSiteConfigRequest{
		Key: ConfigRateTier2Key, Value: strconv.FormatFloat(tier2, 'f', 4, 64), Category: "sales",
	})
	return err
}

// rateForReferral menentukan tier 1 atau tier 2 berdasarkan sudah berapa
// lama partner direferensikan. referredAt nil (seharusnya tidak terjadi,
// selalu diisi bersamaan ReferredBySalesID) fallback ke tier 1.
func (s *Service) rateForReferral(referredAt *time.Time) (rate float64, tier int) {
	tier1, tier2 := s.GetCommissionRates()
	if referredAt == nil {
		return tier1, 1
	}
	threshold := referredAt.AddDate(0, TierThresholdMonths, 0)
	if time.Now().Before(threshold) {
		return tier1, 1
	}
	return tier2, 2
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

	rate, tier := s.rateForReferral(payer.ReferredAt)
	c := &SalesCommission{
		SalesUserID:          *payer.ReferredBySalesID,
		PartnerUserID:        payerUserID,
		PaymentTransactionID: paymentTransactionID,
		OrderID:              orderID,
		SubjectType:          subjectType,
		Tier:                 tier,
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
