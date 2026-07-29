DROP INDEX IF EXISTS idx_events_destination_id;

ALTER TABLE events DROP COLUMN IF EXISTS destination_id;
