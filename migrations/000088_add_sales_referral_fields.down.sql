DROP INDEX IF EXISTS idx_users_referred_by_sales;
ALTER TABLE users DROP COLUMN IF EXISTS referred_by_sales_id;

DROP INDEX IF EXISTS idx_users_referral_code;
ALTER TABLE users DROP COLUMN IF EXISTS referral_code;
