-- Redo of migration 000073_add_bilingual_fields_to_house_ads.
-- The original file was named without the required ".up.sql" suffix, so
-- golang-migrate silently skipped it and prod never received the bilingual
-- house_ads columns. IF NOT EXISTS makes this safe to run even if the
-- columns were already added manually.
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS headline_en  VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS subline_en   TEXT         NOT NULL DEFAULT '';
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS cta_label_en VARCHAR(100) NOT NULL DEFAULT '';
