INSERT INTO permissions (name, created_at, updated_at) VALUES
    ('partner_application.apply', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('partner_application.review', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('partner_application.manage_all', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE p.name = 'partner_application.apply'
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name IN ('admin', 'superadmin')
  AND p.name IN ('partner_application.review', 'partner_application.manage_all')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

