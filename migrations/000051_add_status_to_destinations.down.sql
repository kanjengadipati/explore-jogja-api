DROP INDEX IF EXISTS idx_destinations_status;
ALTER TABLE destinations DROP COLUMN IF EXISTS status;
