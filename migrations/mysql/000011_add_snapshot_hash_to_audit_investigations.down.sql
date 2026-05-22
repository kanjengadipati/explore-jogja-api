DROP INDEX idx_audit_investigations_snapshot_hash ON audit_investigations;

ALTER TABLE audit_investigations DROP COLUMN snapshot_hash;
