-- ============================================================
-- Manual seed: partner role + permissions
-- Run this in production where auto-seed is disabled.
-- Idempotent — safe to re-run (uses ON CONFLICT DO NOTHING).
-- ============================================================

-- 1. Role: partner
INSERT INTO roles (name) VALUES ('partner') ON CONFLICT (name) DO NOTHING;

-- 2. New permissions
INSERT INTO permissions (name) VALUES
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

-- 3. Role → permission mappings
-- Helper: only insert if the pair doesn't already exist.

-- partner role permissions
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'partner'
  AND p.name IN (
      'dashboard.view', 'session.read',
      'partner.read_own', 'partner.create_own', 'partner.update_own', 'partner.delete_own',
      'promotion.manage_own', 'review.reply_own'
  )
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- superadmin: add new partner permissions (existing permissions stay untouched)
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND p.name IN (
      'partner.read_own', 'partner.create_own', 'partner.update_own', 'partner.delete_own',
      'partner.read_all', 'partner.approve', 'partner.reject', 'partner.suspend',
      'partner.manage_all', 'promotion.manage_own', 'review.reply_own', 'ads.manage'
  )
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- admin: add partner management permissions
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN (
      'partner.read_all', 'partner.approve', 'partner.reject',
      'partner.suspend', 'partner.manage_all', 'ads.manage'
  )
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );
