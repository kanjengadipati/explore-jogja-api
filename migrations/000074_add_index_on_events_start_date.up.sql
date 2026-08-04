-- Support the ORDER BY start_date DESC in the event search/list queries.
CREATE INDEX IF NOT EXISTS idx_events_start_date ON events (start_date);
