DROP INDEX IF EXISTS idx_events_content_score;
ALTER TABLE events DROP COLUMN IF EXISTS content_verdict;
ALTER TABLE events DROP COLUMN IF EXISTS content_score;
