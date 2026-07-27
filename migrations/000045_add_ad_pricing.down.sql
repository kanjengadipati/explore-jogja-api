DROP INDEX IF EXISTS idx_ad_campaigns_payment_status;
DROP INDEX IF EXISTS idx_partners_sponsor_payment_status;

ALTER TABLE partners
    DROP COLUMN IF EXISTS sponsor_price,
    DROP COLUMN IF EXISTS sponsor_price_currency,
    DROP COLUMN IF EXISTS sponsor_payment_status;

ALTER TABLE ad_campaigns
    DROP COLUMN IF EXISTS price_amount,
    DROP COLUMN IF EXISTS price_currency,
    DROP COLUMN IF EXISTS payment_status;
