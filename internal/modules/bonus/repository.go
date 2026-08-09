package bonus

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// --- sales_bonuses ---
	CreateBonus(b *SalesBonus) error
	FindByID(id uint) (*SalesBonus, error)
	UpdateStatus(id uint, status Status) error
	FindOnboardingByTenant(salesUserID, tenantUserID uint) (*SalesBonus, error)
	FindMilestoneTier(salesUserID uint, period string, metric BonusMetric, tier int) (*SalesBonus, error)
	VoidOnboardingForTenant(salesUserID, tenantUserID uint) error
	ListBySales(salesUserID uint, limit, offset int) ([]SalesBonus, int64, error)
	ListAll(limit, offset int) ([]SalesBonus, int64, error)
	ListSalesUsers(salesUserIDs []uint) (map[uint]SalesUserInfo, error)

	// Counters fed by the payment settlement — read payment_transactions + users directly.
	CountPaidTransactionsExcluding(payerUserID, excludeTxID uint) (int64, error)
	CountActivatedTenantsBySales(salesUserID uint, period string) (int64, error)
	CountPaidTransactionsBySales(salesUserID uint, period string) (int64, error)

	// --- bonus_rules ---
	ListActiveRules(now time.Time) ([]BonusRule, error)
	ListRules() ([]BonusRule, error)
	CreateRule(r *BonusRule) error
	UpdateRule(r *BonusRule) error
	DeleteRule(id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateBonus(b *SalesBonus) error {
	return r.db.Create(b).Error
}

func (r *repository) FindByID(id uint) (*SalesBonus, error) {
	var b SalesBonus
	err := r.db.First(&b, id).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) UpdateStatus(id uint, status Status) error {
	return r.db.Model(&SalesBonus{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) FindOnboardingByTenant(salesUserID, tenantUserID uint) (*SalesBonus, error) {
	var b SalesBonus
	err := r.db.Where("sales_user_id = ? AND tenant_user_id = ? AND type = ?",
		salesUserID, tenantUserID, BonusTypeOnboarding).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) FindMilestoneTier(salesUserID uint, period string, metric BonusMetric, tier int) (*SalesBonus, error) {
	var b SalesBonus
	err := r.db.Where("sales_user_id = ? AND type = ? AND period = ? AND metric = ? AND tier = ?",
		salesUserID, BonusTypeMilestone, period, metric, tier).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) VoidOnboardingForTenant(salesUserID, tenantUserID uint) error {
	return r.db.Model(&SalesBonus{}).
		Where("sales_user_id = ? AND tenant_user_id = ? AND type = ? AND status != ?",
			salesUserID, tenantUserID, BonusTypeOnboarding, StatusVoided).
		Update("status", StatusVoided).Error
}

func (r *repository) ListBySales(salesUserID uint, limit, offset int) ([]SalesBonus, int64, error) {
	var total int64
	r.db.Model(&SalesBonus{}).Where("sales_user_id = ?", salesUserID).Count(&total)

	var items []SalesBonus
	err := r.db.Where("sales_user_id = ?", salesUserID).
		Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *repository) ListAll(limit, offset int) ([]SalesBonus, int64, error) {
	var total int64
	r.db.Model(&SalesBonus{}).Count(&total)

	var items []SalesBonus
	err := r.db.Order("created_at desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (r *repository) ListSalesUsers(salesUserIDs []uint) (map[uint]SalesUserInfo, error) {
	result := make(map[uint]SalesUserInfo, len(salesUserIDs))
	if len(salesUserIDs) == 0 {
		return result, nil
	}

	type row struct {
		ID    uint
		Name  string
		Email string
	}
	var rows []row
	err := r.db.Table("users").
		Select("id, name, email").
		Where("id IN ?", salesUserIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, rw := range rows {
		result[rw.ID] = SalesUserInfo{Name: rw.Name, Email: rw.Email}
	}
	return result, nil
}

// CountPaidTransactionsExcluding reports whether the payer has other paid
// transactions besides the given one. A zero count means this is the tenant's
// first paid transaction (the onboarding trigger).
func (r *repository) CountPaidTransactionsExcluding(payerUserID, excludeTxID uint) (int64, error) {
	var count int64
	err := r.db.Table("payment_transactions").
		Where("created_by_user_id = ? AND status = ? AND id != ?", payerUserID, "paid", excludeTxID).
		Count(&count).Error
	return count, err
}

// CountActivatedTenantsBySales counts distinct tenants whose FIRST paid
// transaction happened inside the given calendar month, referred by the sales
// user. COALESCE covers the poll-fallback path where paid_at is never set.
func (r *repository) CountActivatedTenantsBySales(salesUserID uint, period string) (int64, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM (
			SELECT pt.created_by_user_id, MIN(COALESCE(pt.paid_at, pt.updated_at)) AS first_paid
			FROM payment_transactions pt
			WHERE pt.status = 'paid' AND pt.created_by_user_id IS NOT NULL
			GROUP BY pt.created_by_user_id
		) f
		JOIN users u ON u.id = f.created_by_user_id
		WHERE u.referred_by_sales_id = ?
		  AND to_char(f.first_paid, 'YYYY-MM') = ?`,
		salesUserID, period).Scan(&count).Error
	return count, err
}

// CountPaidTransactionsBySales counts all paid transactions inside the given
// calendar month whose payer is referred by the sales user.
func (r *repository) CountPaidTransactionsBySales(salesUserID uint, period string) (int64, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM payment_transactions pt
		JOIN users u ON u.id = pt.created_by_user_id
		WHERE pt.status = 'paid'
		  AND u.referred_by_sales_id = ?
		  AND to_char(COALESCE(pt.paid_at, pt.updated_at), 'YYYY-MM') = ?`,
		salesUserID, period).Scan(&count).Error
	return count, err
}

// ListActiveRules returns rules that are active AND whose effective window
// contains `now` (nil windows are unbounded).
func (r *repository) ListActiveRules(now time.Time) ([]BonusRule, error) {
	var rules []BonusRule
	err := r.db.
		Where("is_active = ? AND (effective_from IS NULL OR effective_from <= ?) AND (effective_until IS NULL OR effective_until >= ?)",
			true, now, now).
		Order("type asc, metric asc, tier asc").
		Find(&rules).Error
	return rules, err
}

func (r *repository) ListRules() ([]BonusRule, error) {
	var rules []BonusRule
	err := r.db.Order("type asc, metric asc, tier asc").Find(&rules).Error
	return rules, err
}

func (r *repository) CreateRule(rule *BonusRule) error {
	return r.db.Create(rule).Error
}

func (r *repository) UpdateRule(rule *BonusRule) error {
	return r.db.Save(rule).Error
}

func (r *repository) DeleteRule(id uint) error {
	res := r.db.Delete(&BonusRule{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("bonus rule not found")
	}
	return nil
}
