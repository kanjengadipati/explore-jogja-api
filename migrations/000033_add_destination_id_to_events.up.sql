ALTER TABLE events ADD COLUMN IF NOT EXISTS destination_id VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_events_destination_id ON events(destination_id);
