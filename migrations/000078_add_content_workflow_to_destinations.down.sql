DROP INDEX IF EXISTS idx_destinations_content_status;
ALTER TABLE destinations DROP COLUMN IF EXISTS template_variant;
ALTER TABLE destinations DROP COLUMN IF EXISTS content_status;
