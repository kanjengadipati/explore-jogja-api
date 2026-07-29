-- Link promotions to partner listings. Nullable — generic promos (no partner)
-- remain valid and continue showing on the public promotions feed.

ALTER TABLE promotions
    ADD COLUMN IF NOT EXISTS partner_id VARCHAR(100) REFERENCES partners(external_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_promotions_partner_id ON promotions(partner_id);
