ALTER TABLE partners DROP CONSTRAINT IF EXISTS chk_partners_status;

DROP INDEX IF EXISTS idx_partners_owner_user_id;
DROP INDEX IF EXISTS idx_partners_status;

ALTER TABLE partners
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS submitted_at,
    DROP COLUMN IF EXISTS rejection_reason,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS owner_user_id;
