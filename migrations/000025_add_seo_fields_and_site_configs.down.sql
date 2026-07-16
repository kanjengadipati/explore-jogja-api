DROP TABLE IF EXISTS site_configs;
ALTER TABLE destinations DROP COLUMN IF EXISTS seo_title;
ALTER TABLE destinations DROP COLUMN IF EXISTS seo_keywords;
ALTER TABLE destinations DROP COLUMN IF EXISTS seo_description;
ALTER TABLE destinations DROP COLUMN IF EXISTS og_image_url;
