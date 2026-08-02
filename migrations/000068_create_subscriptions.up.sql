CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    business_id INTEGER NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL DEFAULT 'free',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_period_end TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS chk_subscriptions_plan;
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_plan
    CHECK (plan IN ('free', 'pro', 'business_plus', 'enterprise'));

ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS chk_subscriptions_status;
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_status
    CHECK (status IN ('active', 'past_due', 'canceled'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_business_id ON subscriptions(business_id);

-- Every existing/backfilled business gets a Free plan by default so feature-gate
-- checks never hit a missing row.
INSERT INTO subscriptions (external_id, business_id, plan, status)
SELECT 'sub_' || b.external_id, b.id, 'free', 'active'
FROM businesses b
WHERE NOT EXISTS (SELECT 1 FROM subscriptions s WHERE s.business_id = b.id);

-- Businesses that were already paying for sponsorship/ads get Pro by default —
-- adjust the plan name/threshold with the actual product team before running in prod.
UPDATE subscriptions s
SET plan = 'pro'
FROM businesses b
JOIN partners p ON p.external_id = b.legacy_partner_external_id
WHERE s.business_id = b.id
  AND p.is_sponsored = true
  AND p.sponsor_payment_status = 'paid';
