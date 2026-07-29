INSERT INTO permissions (name)
VALUES
  ('partner.read_own'),
  ('partner.create_own'),
  ('partner.update_own'),
  ('partner.delete_own'),
  ('partner.read_all'),
  ('partner.approve'),
  ('partner.reject'),
  ('partner.suspend'),
  ('partner.manage_all'),
  ('promotion.manage_own'),
  ('review.reply_own'),
  ('ads.manage')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'partner'
  AND p.name IN (
    'dashboard.view',
    'session.read',
    'partner.read_own',
    'partner.create_own',
    'partner.update_own',
    'partner.delete_own',
    'promotion.manage_own',
    'review.reply_own'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM role_permissions rp
      WHERE rp.role_id = r.id
        AND rp.permission = p.name
        AND rp.deleted_at IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND p.name IN (
    'partner.read_own',
    'partner.create_own',
    'partner.update_own',
    'partner.delete_own',
    'partner.read_all',
    'partner.approve',
    'partner.reject',
    'partner.suspend',
    'partner.manage_all',
    'promotion.manage_own',
    'review.reply_own',
    'ads.manage'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM role_permissions rp
      WHERE rp.role_id = r.id
        AND rp.permission = p.name
        AND rp.deleted_at IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN (
    'partner.read_all',
    'partner.approve',
    'partner.reject',
    'partner.suspend',
    'partner.manage_all',
    'ads.manage'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM role_permissions rp
      WHERE rp.role_id = r.id
        AND rp.permission = p.name
        AND rp.deleted_at IS NULL
  );
