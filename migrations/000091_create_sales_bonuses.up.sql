-- Sales bonus scheme (see design doc):
--   - onboarding: one-time flat bonus per tenant activated (first paid transaction),
--     triggered by payment settlement, voided if the tenant's first transaction is refunded.
--   - milestone: per calendar-month (period 'YYYY-MM'), tiered-additive bonus rows,
--     one per tier reached. Recalculated in real time on each settlement (no cron).

CREATE TABLE IF NOT EXISTS sales_bonuses (
    id BIGSERIAL PRIMARY KEY,
    sales_user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,             -- onboarding | milestone
    tenant_user_id BIGINT,                 -- only for onboarding
    period VARCHAR(7),                     -- 'YYYY-MM' only for milestone
    metric VARCHAR(20),                    -- tenant | transaction (milestone only)
    tier INT,                              -- only for milestone
    amount DOUBLE PRECISION NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sales_bonuses_sales_user ON sales_bonuses(sales_user_id);
CREATE INDEX IF NOT EXISTS idx_sales_bonuses_type ON sales_bonuses(type);
CREATE INDEX IF NOT EXISTS idx_sales_bonuses_period ON sales_bonuses(period);
CREATE INDEX IF NOT EXISTS idx_sales_bonuses_status ON sales_bonuses(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_bonuses_onboarding_tenant
    ON sales_bonuses(sales_user_id, tenant_user_id)
    WHERE type = 'onboarding' AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_bonuses_milestone_tier
    ON sales_bonuses(sales_user_id, period, metric, tier)
    WHERE type = 'milestone' AND deleted_at IS NULL;

-- Admin-editable bonus rules (nominal for onboarding, tier thresholds for milestone).
CREATE TABLE IF NOT EXISTS bonus_rules (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL,             -- onboarding | milestone
    metric VARCHAR(20) NOT NULL DEFAULT 'tenant', -- tenant | transaction (milestone only)
    tier INT,                              -- milestone only
    threshold INT,                         -- milestone only: count needed to unlock the tier
    amount DOUBLE PRECISION NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    effective_from DATE,
    effective_until DATE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bonus_rules_type ON bonus_rules(type);
CREATE INDEX IF NOT EXISTS idx_bonus_rules_active ON bonus_rules(is_active);
