INSERT INTO permissions (name, created_at, updated_at) VALUES
    ('partner_application.apply', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('partner_application.review', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('partner_application.manage_all', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;
