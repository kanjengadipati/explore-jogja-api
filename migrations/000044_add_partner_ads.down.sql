DROP TABLE IF EXISTS ad_campaigns;

DROP INDEX IF EXISTS idx_partners_target_dest_ids;
DROP INDEX IF EXISTS idx_partners_is_sponsored;

ALTER TABLE partners
    DROP COLUMN IF EXISTS click_count,
    DROP COLUMN IF EXISTS impression_count,
    DROP COLUMN IF EXISTS target_dest_ids,
    DROP COLUMN IF EXISTS sponsor_end_at,
    DROP COLUMN IF EXISTS sponsor_start_at,
    DROP COLUMN IF EXISTS sponsor_tier,
    DROP COLUMN IF EXISTS is_sponsored;
