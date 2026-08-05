ALTER TABLE destinations ADD COLUMN IF NOT EXISTS content_score   INTEGER      NOT NULL DEFAULT 0;
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS content_verdict VARCHAR(20)  NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_destinations_content_score ON destinations(content_score);
