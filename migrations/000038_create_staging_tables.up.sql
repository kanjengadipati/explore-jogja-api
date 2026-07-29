CREATE TABLE IF NOT EXISTS staging_destinations (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    source TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    latitude TEXT,
    longitude TEXT,
    address TEXT,
    category TEXT,
    images TEXT,
    raw_data TEXT,
    status TEXT DEFAULT 'pending'
);

CREATE INDEX IF NOT EXISTS idx_staging_destinations_status ON staging_destinations(status);
CREATE INDEX IF NOT EXISTS idx_staging_destinations_deleted_at ON staging_destinations(deleted_at);

CREATE TABLE IF NOT EXISTS staging_events (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    source TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    location TEXT,
    raw_data TEXT,
    status TEXT DEFAULT 'pending'
);

CREATE INDEX IF NOT EXISTS idx_staging_events_status ON staging_events(status);
CREATE INDEX IF NOT EXISTS idx_staging_events_deleted_at ON staging_events(deleted_at);
