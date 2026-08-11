ALTER TABLE events ADD COLUMN IF NOT EXISTS content_score   INTEGER      NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN IF NOT EXISTS content_verdict VARCHAR(20)  NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_events_content_score ON events(content_score);
