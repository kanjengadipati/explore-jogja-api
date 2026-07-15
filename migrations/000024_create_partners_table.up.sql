CREATE TABLE partners (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    location VARCHAR(255),
    address VARCHAR(500),
    image VARCHAR(500),
    rating DECIMAL(3,1) DEFAULT 0,
    price VARCHAR(255),
    distance VARCHAR(100),
    phone VARCHAR(50),
    website VARCHAR(500),
    latitude DECIMAL(10,6),
    longitude DECIMAL(10,6),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_partners_external_id ON partners(external_id);
CREATE INDEX idx_partners_category ON partners(category);
CREATE INDEX idx_partners_deleted_at ON partners(deleted_at);
