-- Phase 4: promotions get a real FK to businesses.
-- The old partner_id is a clean rename (values carry over as-is) into
-- legacy_partner_external_id, and business_external_id becomes the real FK.
ALTER TABLE promotions RENAME COLUMN partner_id TO legacy_partner_external_id;
ALTER TABLE promotions ADD COLUMN IF NOT EXISTS business_external_id VARCHAR(100)
    REFERENCES businesses(external_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_promotions_business_external_id ON promotions(business_external_id);

-- Backfill: link each promotion to its business mirror via the carried-over
-- legacy partner external id. Promotions without a matching business stay NULL.
UPDATE promotions p
SET business_external_id = b.external_id
FROM businesses b
WHERE b.legacy_partner_external_id = p.legacy_partner_external_id
  AND p.legacy_partner_external_id IS NOT NULL;

-- Verify before dropping the old column (manual step — do NOT run in migration):
--   SELECT COUNT(*) FROM promotions WHERE legacy_partner_external_id IS NOT NULL AND business_external_id IS NULL; -- expect 0
-- Then, once verified:
--   ALTER TABLE promotions DROP COLUMN legacy_partner_external_id;
