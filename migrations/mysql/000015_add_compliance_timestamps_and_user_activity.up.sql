ALTER TABLE users
ADD COLUMN last_login_at DATETIME,
ADD COLUMN last_password_change DATETIME;

UPDATE users
SET last_password_change = COALESCE(password_updated_at, created_at)
WHERE last_password_change IS NULL;

ALTER TABLE email_verification_tokens
ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN deleted_at DATETIME;

ALTER TABLE permissions
ADD COLUMN created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN deleted_at DATETIME;

ALTER TABLE roles
ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN deleted_at DATETIME;

ALTER TABLE role_permissions
ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN deleted_at DATETIME;

ALTER TABLE social_accounts
ADD COLUMN deleted_at DATETIME;

CREATE INDEX idx_users_last_login_at ON users (last_login_at DESC);
CREATE INDEX idx_users_last_password_change ON users (last_password_change DESC);
CREATE INDEX idx_email_verification_tokens_deleted_at ON email_verification_tokens (deleted_at);
CREATE INDEX idx_permissions_deleted_at ON permissions (deleted_at);
CREATE INDEX idx_roles_deleted_at ON roles (deleted_at);
CREATE INDEX idx_role_permissions_deleted_at ON role_permissions (deleted_at);
CREATE INDEX idx_social_accounts_deleted_at ON social_accounts (deleted_at);
