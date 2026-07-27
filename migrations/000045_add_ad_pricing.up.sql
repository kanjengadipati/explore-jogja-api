-- Flat-fee pricing for sponsored partners and ad campaigns. Admin sets the
-- price manually per campaign/partner; there is no automatic CPM/CPC billing
-- engine. Payment status is tracked manually without payment gateway integration.

ALTER TABLE partners
    ADD COLUMN IF NOT EXISTS sponsor_price NUMERIC(14,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sponsor_price_currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    ADD COLUMN IF NOT EXISTS sponsor_payment_status VARCHAR(20) NOT NULL DEFAULT 'pending';

ALTER TABLE ad_campaigns
    ADD COLUMN IF NOT EXISTS price_amount NUMERIC(14,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS price_currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    ADD COLUMN IF NOT EXISTS payment_status VARCHAR(20) NOT NULL DEFAULT 'pending';

CREATE INDEX IF NOT EXISTS idx_ad_campaigns_payment_status ON ad_campaigns(payment_status);
CREATE INDEX IF NOT EXISTS idx_partners_sponsor_payment_status ON partners(sponsor_payment_status);
