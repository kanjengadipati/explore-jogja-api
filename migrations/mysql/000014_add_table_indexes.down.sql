DROP INDEX idx_error_analyses_severity ON error_analyses;
DROP INDEX idx_error_analyses_error_type ON error_analyses;
DROP INDEX idx_error_analyses_created_at ON error_analyses;
DROP INDEX idx_error_analyses_deleted_at ON error_analyses;

DROP INDEX idx_audit_investigations_created_by_created_at ON audit_investigations;
DROP INDEX idx_audit_investigations_ai_model ON audit_investigations;
DROP INDEX idx_audit_investigations_ai_provider ON audit_investigations;
DROP INDEX idx_audit_investigations_status ON audit_investigations;

DROP INDEX idx_audit_logs_resource_status_created_at ON audit_logs;
DROP INDEX idx_audit_logs_actor_created_at ON audit_logs;
DROP INDEX idx_audit_logs_resource_id ON audit_logs;
DROP INDEX idx_audit_logs_status ON audit_logs;
DROP INDEX idx_audit_logs_created_at ON audit_logs;

DROP INDEX idx_social_accounts_user_provider ON social_accounts;
DROP INDEX idx_social_accounts_user_id ON social_accounts;

DROP INDEX idx_role_permissions_role_permission ON role_permissions;
DROP INDEX idx_role_permissions_permission ON role_permissions;
DROP INDEX idx_role_permissions_role_id ON role_permissions;

DROP INDEX idx_refresh_tokens_expired_at ON refresh_tokens;
DROP INDEX idx_refresh_tokens_deleted_at ON refresh_tokens;
DROP INDEX idx_refresh_tokens_user_device ON refresh_tokens;
DROP INDEX idx_refresh_tokens_user_created_at ON refresh_tokens;

DROP INDEX idx_email_verification_tokens_expires_at ON email_verification_tokens;
DROP INDEX idx_email_verification_tokens_token ON email_verification_tokens;
DROP INDEX idx_email_verification_tokens_user_id ON email_verification_tokens;

DROP INDEX idx_users_created_at ON users;
DROP INDEX idx_users_deleted_at ON users;
DROP INDEX idx_users_role_id ON users;
DROP INDEX idx_users_role ON users;
