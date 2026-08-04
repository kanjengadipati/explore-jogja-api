-- migration: 000073_add_bilingual_fields_to_house_ads
-- Adds English-language variants of content fields to house_ads table.
-- Existing rows will have empty strings — admin should fill them in
-- via the house-ads management page.

-- up
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS headline_en  VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS subline_en   TEXT         NOT NULL DEFAULT '';
ALTER TABLE house_ads ADD COLUMN IF NOT EXISTS cta_label_en VARCHAR(100) NOT NULL DEFAULT '';

-- down
ALTER TABLE house_ads DROP COLUMN IF EXISTS headline_en;
ALTER TABLE house_ads DROP COLUMN IF EXISTS subline_en;
ALTER TABLE house_ads DROP COLUMN IF EXISTS cta_label_en;
