ALTER TABLE events ADD COLUMN IF NOT EXISTS title_en VARCHAR(255) DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS description_en TEXT DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255) DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_title_en VARCHAR(255) DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_description TEXT DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_description_en TEXT DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_keywords TEXT DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS seo_keywords_en TEXT DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS og_image_url VARCHAR(500) DEFAULT '';

-- Backfill English columns from existing content.
UPDATE events SET
    title_en = title,
    description_en = description
WHERE title_en = '' OR title_en IS NULL;
