ALTER TABLE user_destinations ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
CREATE INDEX IF NOT EXISTS idx_user_destinations_deleted_at ON user_destinations(deleted_at);
