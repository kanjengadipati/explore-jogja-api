INSERT INTO permissions (name, created_at, updated_at)
VALUES ('payment.manage', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;
