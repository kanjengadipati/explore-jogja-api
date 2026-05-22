DROP INDEX idx_social_accounts_deleted_at ON social_accounts;
DROP INDEX idx_role_permissions_deleted_at ON role_permissions;
DROP INDEX idx_roles_deleted_at ON roles;
DROP INDEX idx_permissions_deleted_at ON permissions;
DROP INDEX idx_email_verification_tokens_deleted_at ON email_verification_tokens;
DROP INDEX idx_users_last_password_change ON users;
DROP INDEX idx_users_last_login_at ON users;

ALTER TABLE social_accounts DROP COLUMN deleted_at;

ALTER TABLE role_permissions
DROP COLUMN deleted_at,
DROP COLUMN updated_at;

ALTER TABLE roles
DROP COLUMN deleted_at,
DROP COLUMN updated_at;

ALTER TABLE permissions
DROP COLUMN deleted_at,
DROP COLUMN updated_at,
DROP COLUMN created_at;

ALTER TABLE email_verification_tokens
DROP COLUMN deleted_at,
DROP COLUMN updated_at;

ALTER TABLE users
DROP COLUMN last_password_change,
DROP COLUMN last_login_at;
