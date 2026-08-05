-- content_status: tracks AI-generated copy state (draft/review/published/flagged/rejected)
-- template_variant: which of the 3 AI prompt strategies was used
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS content_status   VARCHAR(20) NOT NULL DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS template_variant VARCHAR(30) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_destinations_content_status ON destinations(content_status) WHERE content_status != '';
