CREATE TABLE error_messages (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  error_code VARCHAR(100) NOT NULL UNIQUE,
  error_type VARCHAR(50) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  generic_message TEXT NOT NULL,
  ai_message TEXT,
  ai_suggestions JSON,
  ai_generated_at DATETIME,
  description TEXT,
  is_sensitive BOOLEAN DEFAULT false,
  should_expose_details BOOLEAN DEFAULT false,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_error_code ON error_messages(error_code);
CREATE INDEX idx_error_type ON error_messages(error_type);

CREATE TABLE error_message_feedback (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED,
  error_code VARCHAR(100) NOT NULL,
  was_helpful BOOLEAN,
  clarity_rating INTEGER,
  action_taken VARCHAR(100),
  comments TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_error_message_feedback_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_feedback_error_code ON error_message_feedback(error_code);
CREATE INDEX idx_feedback_user_id ON error_message_feedback(user_id);

CREATE TABLE error_analytics (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  error_code VARCHAR(100) NOT NULL UNIQUE,
  error_type VARCHAR(50) NOT NULL,
  occurrence_count INTEGER DEFAULT 1,
  last_occurred DATETIME,
  ai_message_version INTEGER DEFAULT 1,
  avg_helpfulness_rating FLOAT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_analytics_error_code ON error_analytics(error_code);
CREATE INDEX idx_analytics_last_occurred ON error_analytics(last_occurred);

CREATE TABLE error_context_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT UNSIGNED,
  error_code VARCHAR(100) NOT NULL,
  error_context JSON,
  request_path VARCHAR(255),
  request_method VARCHAR(10),
  ip_address VARCHAR(255),
  device_id VARCHAR(255),
  status_code INTEGER,
  response_sent JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_context_user_id ON error_context_logs(user_id);
CREATE INDEX idx_context_error_code ON error_context_logs(error_code);
CREATE INDEX idx_context_created_at ON error_context_logs(created_at);
