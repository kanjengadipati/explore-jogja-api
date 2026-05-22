CREATE TABLE error_analyses (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    error_message TEXT NOT NULL,
    error_type VARCHAR(255) NOT NULL,
    root_cause TEXT,
    affected_components TEXT,
    recommended_action TEXT,
    severity VARCHAR(50)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
