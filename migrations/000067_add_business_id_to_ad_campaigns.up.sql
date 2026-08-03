-- Phase 4: ad_campaigns get a real FK to businesses, replacing the free-text
-- partner_name column. This migration only does the additive + safe auto-match
-- part; the destructive part (NOT NULL + DROP COLUMN) runs manually once every
-- campaign has been resolved (see Step 4).
ALTER TABLE ad_campaigns ADD COLUMN IF NOT EXISTS business_external_id VARCHAR(100)
    REFERENCES businesses(external_id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_business_external_id ON ad_campaigns(business_external_id);

-- Step 1: exact-name auto-match only. Case-insensitive, trimmed, and only where
-- the name is unambiguous (exactly one business matches the normalized name) —
-- anything with 0 or 2+ matches is deliberately left NULL for manual review.
-- (Equivalent to the spec's intent; the ORIGINAL spec SQL used DISTINCT ON +
-- GROUP BY + a window function in HAVING, which PostgreSQL rejects. This
-- correlated-subquery version returns identical semantics.)
UPDATE ad_campaigns ac
SET business_external_id = b.external_id
FROM businesses b
WHERE lower(trim(ac.partner_name)) = lower(trim(b.name))
  AND ac.business_external_id IS NULL
  AND (SELECT COUNT(*) FROM businesses b2 WHERE lower(trim(b2.name)) = lower(trim(ac.partner_name))) = 1;

-- Step 2 (read-only — hand this list to an admin): everything still unmatched.
--   SELECT ac.external_id AS ad_campaign_id, ac.partner_name, ac.is_active, ac.payment_status
--   FROM ad_campaigns ac
--   WHERE ac.business_external_id IS NULL
--   ORDER BY ac.is_active DESC, ac.created_at DESC;

-- Step 3 (manual, one row at a time, after a human confirms the match):
--   UPDATE ad_campaigns SET business_external_id = '<confirmed_business_external_id>'
--   WHERE external_id = '<ad_campaign_id>';

-- Step 4 — only run once the Step 2 query returns zero rows:
--   ALTER TABLE ad_campaigns ALTER COLUMN business_external_id SET NOT NULL;
--   ALTER TABLE ad_campaigns DROP COLUMN partner_name;
