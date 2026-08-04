ALTER TABLE destinations DROP COLUMN IF EXISTS last_scraped_at;
ALTER TABLE events DROP COLUMN IF EXISTS last_scraped_at;
