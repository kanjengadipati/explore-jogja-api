CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    start_date VARCHAR(50),
    end_date VARCHAR(50),
    image_url VARCHAR(1000),
    category VARCHAR(100),
    status VARCHAR(50),
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    max_attendees INTEGER DEFAULT 0,
    ticket_price VARCHAR(255),
    organizer VARCHAR(255),
    highlights JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_external_id ON events(external_id);
CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
