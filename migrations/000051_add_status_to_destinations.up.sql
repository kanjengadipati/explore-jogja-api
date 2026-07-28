ALTER TABLE destinations ADD COLUMN status VARCHAR(50) DEFAULT 'published';
CREATE INDEX idx_destinations_status ON destinations(status);
