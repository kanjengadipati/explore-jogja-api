CREATE TABLE audit_investigations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    created_by_user_id BIGINT UNSIGNED NULL,
    action TEXT NOT NULL,
    resource TEXT NOT NULL,
    status TEXT NOT NULL,
    actor_user_id BIGINT UNSIGNED NULL,
    search TEXT NOT NULL,
    date_from DATETIME NULL,
    date_to DATETIME NULL,
    limit_value INTEGER NOT NULL DEFAULT 50,
    log_count INTEGER NOT NULL DEFAULT 0,
    ai_provider TEXT NOT NULL,
    ai_model TEXT NOT NULL,
    summary TEXT NOT NULL,
    timeline_json JSON NOT NULL,
    suspicious_signals_json JSON NOT NULL,
    recommendations_json JSON NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_audit_investigations_deleted_at ON audit_investigations (deleted_at);
CREATE INDEX idx_audit_investigations_created_at ON audit_investigations (created_at DESC);
CREATE INDEX idx_audit_investigations_created_by_user_id ON audit_investigations (created_by_user_id);
CREATE INDEX idx_audit_investigations_resource_status ON audit_investigations (resource(100), status(100));
