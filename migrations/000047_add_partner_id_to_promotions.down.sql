DROP INDEX IF EXISTS idx_promotions_partner_id;

ALTER TABLE promotions
    DROP COLUMN IF EXISTS partner_id;
