ALTER TABLE users ADD COLUMN IF NOT EXISTS referral_code VARCHAR(20);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL;

ALTER TABLE users ADD COLUMN IF NOT EXISTS referred_by_sales_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_users_referred_by_sales ON users(referred_by_sales_id);
