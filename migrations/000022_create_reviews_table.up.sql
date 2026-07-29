CREATE TABLE reviews (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    user_id VARCHAR(100),
    destination_id VARCHAR(100),
    user_name VARCHAR(255),
    traveler_type VARCHAR(50),
    rating INTEGER DEFAULT 0,
    comment TEXT,
    images JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'published',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_reviews_external_id ON reviews(external_id);
CREATE INDEX idx_reviews_destination_id ON reviews(destination_id);
CREATE INDEX idx_reviews_user_id ON reviews(user_id);
CREATE INDEX idx_reviews_status ON reviews(status);
CREATE INDEX idx_reviews_traveler_type ON reviews(traveler_type);
CREATE INDEX idx_reviews_deleted_at ON reviews(deleted_at);
