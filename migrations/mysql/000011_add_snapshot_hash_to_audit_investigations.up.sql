ALTER TABLE audit_investigations ADD COLUMN snapshot_hash VARCHAR(255) NOT NULL DEFAULT '';

CREATE INDEX idx_audit_investigations_snapshot_hash
ON audit_investigations (created_by_user_id, snapshot_hash, created_at DESC);
