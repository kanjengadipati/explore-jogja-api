-- destinations.partners is free-text tourism-affiliate data, unrelated to the
-- Partner entity. Rename it now (before the Business migration makes this
-- ambiguous) so ownership-related lookups never confuse the two.
-- Idempotent: only rename if the old column still exists (dirty retry path).
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_schema = current_schema()
                 AND table_name = 'destinations'
                 AND column_name = 'partners') THEN
        ALTER TABLE destinations RENAME COLUMN partners TO affiliate_partners;
    END IF;
END $$;
