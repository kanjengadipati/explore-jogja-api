DROP INDEX IF EXISTS idx_reviews_partner_id;

ALTER TABLE reviews
    DROP COLUMN IF EXISTS replied_by,
    DROP COLUMN IF EXISTS replied_at,
    DROP COLUMN IF EXISTS reply,
    DROP COLUMN IF EXISTS partner_id;
