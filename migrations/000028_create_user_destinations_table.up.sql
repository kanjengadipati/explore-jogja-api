CREATE TABLE user_destinations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination_slug VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    UNIQUE(user_id, destination_slug)
);

CREATE INDEX idx_user_destinations_user_id ON user_destinations(user_id);
CREATE INDEX idx_user_destinations_slug ON user_destinations(destination_slug);
CREATE INDEX idx_user_destinations_deleted_at ON user_destinations(deleted_at);
