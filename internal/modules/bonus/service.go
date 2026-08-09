package bonus

import (
	"errors"
	"strings"
	"time"

	"pleco-api/internal/modules/config"
	"pleco-api/internal/modules/user"
)

type Service struct {
	repo      Repository
	configSvc *config.Service
	userSvc   *user.Service
}

func NewService(repo Repository, configSvc *config.Service, userSvc *user.Service) *Service {
	return &Service{repo: repo, configSvc: configSvc, userSvc: userSvc}
}

// RecordFromTransaction is called on every payment settlement. It:
//  1. records a one-time onboarding bonus when this is the payer's FIRST paid
//     transaction (the "tenant activation" event — not signup), and
//  2. recalculates milestone bonuses for the settlement's calendar month.
//
// Tiered-additive: each milestone tier reached pays its own amount, and rows
// are created for every qualifying tier. Safe to call more than once for the
// same transaction — existence checks + unique partial indexes make it
// idempotent across webhook retries.
func (s *Service) RecordFromTransaction(payerUserID, paymentTransactionID uint, paidAt time.Time) error {
	payer, err := s.userSvc.UserRepo.FindByID(payerUserID)
	if err != nil || payer.ReferredBySalesID == nil {
		return nil // payer not found, or not referred by any sales user — no bonus owed
	}
	salesUserID := *payer.ReferredBySalesID

	firstPaidCount, err := s.repo.CountPaidTransactionsExcluding(payerUserID, paymentTransactionID)
	if err != nil {
		return err
	}
	if firstPaidCount == 0 {
		if err := s.recordOnboarding(salesUserID, payerUserID); err != nil {
			return err
		}
	}

	return s.recordMilestones(salesUserID, paidAt)
}

func (s *Service) recordOnboarding(salesUserID, tenantUserID uint) error {
	if existing, err := s.repo.FindOnboardingByTenant(salesUserID, tenantUserID); err == nil && existing != nil {
		return nil // already recorded
	}

	rules, err := s.repo.ListActiveRules(time.Now())
	if err != nil {
		return err
	}
	var rule *BonusRule
	for i := range rules {
		if rules[i].Type == BonusTypeOnboarding {
			rule = &rules[i]
			break
		}
	}
	if rule == nil {
		return nil // no onboarding bonus configured — nothing to record
	}

	b := &SalesBonus{
		SalesUserID:  salesUserID,
		Type:         BonusTypeOnboarding,
		TenantUserID: &tenantUserID,
		Amount:       rule.Amount,
		Status:       StatusPending,
	}
	if err := s.repo.CreateBonus(b); err != nil {
		if isDuplicateKey(err) {
			return nil // another webhook call already recorded it
		}
		return err
	}
	return nil
}

// recordMilestones recomputes both metrics (activated tenants and settled
// transactions) for the sales user in the settlement's calendar month and
// creates a pending bonus row for every tier whose threshold is now reached
// and that hasn't been rewarded yet.
func (s *Service) recordMilestones(salesUserID uint, paidAt time.Time) error {
	period := paidAt.Format("2006-01")
	rules, err := s.repo.ListActiveRules(paidAt)
	if err != nil {
		return err
	}

	for _, metric := range []BonusMetric{MetricTenant, MetricTransaction} {
		var count int64
		switch metric {
		case MetricTenant:
			count, err = s.repo.CountActivatedTenantsBySales(salesUserID, period)
		case MetricTransaction:
			count, err = s.repo.CountPaidTransactionsBySales(salesUserID, period)
		}
		if err != nil {
			return err
		}

		for i := range rules {
			r := &rules[i]
			if r.Type != BonusTypeMilestone || r.Metric != metric || r.Tier == nil || r.Threshold == nil {
				continue
			}
			if count < int64(*r.Threshold) {
				continue
			}
			if existing, err := s.repo.FindMilestoneTier(salesUserID, period, metric, *r.Tier); err == nil && existing != nil {
				continue // tier already rewarded for this period
			}

			tier := *r.Tier
			b := &SalesBonus{
				SalesUserID: salesUserID,
				Type:        BonusTypeMilestone,
				Period:      &period,
				Metric:      metric,
				Tier:        &tier,
				Amount:      r.Amount,
				Status:      StatusPending,
			}
			if err := s.repo.CreateBonus(b); err != nil {
				if isDuplicateKey(err) {
					continue
				}
				return err
			}
		}
	}
	return nil
}

// VoidOnboardingForTenant voids the tenant's onboarding bonus, e.g. when their
// first transaction is refunded. Not wired to any trigger yet — the payment
// module has no refund flow at the moment — but ready for when one exists.
func (s *Service) VoidOnboardingForTenant(tenantUserID uint) error {
	tenant, err := s.userSvc.UserRepo.FindByID(tenantUserID)
	if err != nil || tenant.ReferredBySalesID == nil {
		return nil
	}
	return s.repo.VoidOnboardingForTenant(*tenant.ReferredBySalesID, tenantUserID)
}

func (s *Service) ListMyBonuses(salesUserID uint, limit, offset int) ([]BonusResponse, int64, error) {
	items, total, err := s.repo.ListBySales(salesUserID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	res := make([]BonusResponse, len(items))
	for i := range items {
		res[i] = toBonusResponse(&items[i])
	}
	return res, total, nil
}

func (s *Service) ListAllBonuses(limit, offset int) ([]BonusResponse, int64, error) {
	items, total, err := s.repo.ListAll(limit, offset)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uint, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].SalesUserID)
	}
	names, err := s.repo.ListSalesUsers(ids)
	if err != nil {
		return nil, 0, err
	}
	res := make([]BonusResponse, len(items))
	for i := range items {
		res[i] = toBonusResponse(&items[i])
		if info, ok := names[items[i].SalesUserID]; ok {
			res[i].SalesUserName = info.Name
			res[i].SalesUserEmail = info.Email
		}
	}
	return res, total, nil
}

func (s *Service) ListBonusRules() ([]BonusRuleResponse, error) {
	rules, err := s.repo.ListRules()
	if err != nil {
		return nil, err
	}
	res := make([]BonusRuleResponse, len(rules))
	for i := range rules {
		res[i] = toBonusRuleResponse(&rules[i])
	}
	return res, nil
}

func (s *Service) CreateBonusRule(req CreateBonusRuleRequest) (*BonusRule, error) {
	rule := &BonusRule{
		Type:      req.Type,
		Metric:    req.Metric,
		Tier:      req.Tier,
		Threshold: req.Threshold,
		Amount:    req.Amount,
		IsActive:  true,
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.EffectiveFrom != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveFrom)
		if err != nil {
			return nil, errors.New("effective_from must be a date in YYYY-MM-DD format")
		}
		rule.EffectiveFrom = &t
	}
	if req.EffectiveUntil != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveUntil)
		if err != nil {
			return nil, errors.New("effective_until must be a date in YYYY-MM-DD format")
		}
		rule.EffectiveUntil = &t
	}
	if err := validateRule(rule); err != nil {
		return nil, err
	}
	if err := s.repo.CreateRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) UpdateBonusRule(id uint, req CreateBonusRuleRequest) (*BonusRule, error) {
	rule, err := s.findRule(id)
	if err != nil {
		return nil, err
	}

	rule.Type = req.Type
	rule.Metric = req.Metric
	rule.Tier = req.Tier
	rule.Threshold = req.Threshold
	rule.Amount = req.Amount
	rule.EffectiveFrom = nil
	rule.EffectiveUntil = nil
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	if req.EffectiveFrom != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveFrom)
		if err != nil {
			return nil, errors.New("effective_from must be a date in YYYY-MM-DD format")
		}
		rule.EffectiveFrom = &t
	}
	if req.EffectiveUntil != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveUntil)
		if err != nil {
			return nil, errors.New("effective_until must be a date in YYYY-MM-DD format")
		}
		rule.EffectiveUntil = &t
	}
	if err := validateRule(rule); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *Service) DeleteBonusRule(id uint) error {
	return s.repo.DeleteRule(id)
}

// MarkStatus is the admin payout action. Only a pending bonus can be marked
// paid (money transferred) or voided (payout cancelled). Paid/voided are final.
func (s *Service) MarkStatus(id uint, status Status) (*BonusResponse, error) {
	b, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("bonus not found")
	}
	if b.Status != StatusPending {
		return nil, errors.New("only pending bonuses can be marked as paid or voided")
	}
	if err := s.repo.UpdateStatus(id, status); err != nil {
		return nil, err
	}
	b.Status = status
	res := toBonusResponse(b)
	return &res, nil
}

func (s *Service) findRule(id uint) (*BonusRule, error) {
	rules, err := s.repo.ListRules()
	if err != nil {
		return nil, err
	}
	for i := range rules {
		if rules[i].ID == id {
			return &rules[i], nil
		}
	}
	return nil, errors.New("bonus rule not found")
}

func validateRule(rule *BonusRule) error {
	switch rule.Type {
	case BonusTypeOnboarding:
		// Onboarding is a flat nominal — no tier/threshold.
	case BonusTypeMilestone:
		if rule.Tier == nil || *rule.Tier <= 0 {
			return errors.New("milestone rules require a tier > 0")
		}
		if rule.Threshold == nil || *rule.Threshold <= 0 {
			return errors.New("milestone rules require a threshold > 0")
		}
		if rule.Metric != MetricTenant && rule.Metric != MetricTransaction {
			return errors.New("milestone metric must be 'tenant' or 'transaction'")
		}
	default:
		return errors.New("rule type must be 'onboarding' or 'milestone'")
	}
	if rule.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if rule.EffectiveFrom != nil && rule.EffectiveUntil != nil && rule.EffectiveFrom.After(*rule.EffectiveUntil) {
		return errors.New("effective_from cannot be after effective_until")
	}
	return nil
}

func isDuplicateKey(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "UNIQUE constraint failed")
}
