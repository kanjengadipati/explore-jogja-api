DROP INDEX IF EXISTS idx_destinations_content_score;
ALTER TABLE destinations DROP COLUMN IF EXISTS content_verdict;
ALTER TABLE destinations DROP COLUMN IF EXISTS content_score;
