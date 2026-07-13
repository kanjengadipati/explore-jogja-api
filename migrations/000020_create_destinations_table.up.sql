CREATE TABLE destinations (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    tagline VARCHAR(500),
    category VARCHAR(100),
    location VARCHAR(255),
    sub_region VARCHAR(100),
    images JSONB DEFAULT '[]',
    rating DECIMAL(3,1) DEFAULT 0,
    review_count INTEGER DEFAULT 0,
    description TEXT,
    story TEXT,
    ticket_price VARCHAR(255),
    opening_hours VARCHAR(255),
    facilities JSONB DEFAULT '[]',
    travel_tips JSONB DEFAULT '[]',
    best_time VARCHAR(255),
    weather JSONB DEFAULT '{}',
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    reviews JSONB DEFAULT '[]',
    partners JSONB DEFAULT '[]',
    faqs JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_destinations_external_id ON destinations(external_id);
CREATE INDEX idx_destinations_category ON destinations(category);
CREATE INDEX idx_destinations_sub_region ON destinations(sub_region);
CREATE INDEX idx_destinations_deleted_at ON destinations(deleted_at);
