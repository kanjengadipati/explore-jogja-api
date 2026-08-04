-- Tracks when the scraper last touched a row so it can tell whether a
-- manual (admin) edit happened since. The scraper only auto-updates a row
-- when its updated_at is not newer than last_scraped_at.
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS last_scraped_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE events ADD COLUMN IF NOT EXISTS last_scraped_at TIMESTAMP WITH TIME ZONE;
