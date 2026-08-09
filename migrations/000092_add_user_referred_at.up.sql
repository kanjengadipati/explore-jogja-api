ALTER TABLE users ADD COLUMN IF NOT EXISTS referred_at TIMESTAMP;

-- Backfill: partners already referred (referred_by_sales_id set) before the
-- tiered commission rollout get their tier countdown started from their actual
-- signup time, so nobody stays on tier 1 (20%) forever.
UPDATE users
   SET referred_at = created_at
 WHERE referred_by_sales_id IS NOT NULL
   AND referred_at IS NULL;
