-- Add English (_en) columns for bilingual content.
-- Existing columns remain as Indonesian (default language).
-- Handler resolves Accept-Language to return the right set.

ALTER TABLE destinations ADD COLUMN IF NOT EXISTS name_en VARCHAR(255) DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS tagline_en VARCHAR(500) DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS description_en TEXT DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS story_en TEXT DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS best_time_en VARCHAR(255) DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS facilities_en JSONB DEFAULT '[]';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS travel_tips_en JSONB DEFAULT '[]';

-- Migrate existing English content from main columns to _en columns
-- (main columns will be overwritten with Indonesian by seed)
UPDATE destinations SET
    name_en = name,
    tagline_en = tagline,
    description_en = description,
    story_en = story,
    best_time_en = best_time,
    facilities_en = facilities,
    travel_tips_en = travel_tips
WHERE name_en = '' OR name_en IS NULL;
