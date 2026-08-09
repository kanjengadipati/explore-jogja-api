-- ============================================================
-- Manual seed: sales role + commission permissions
-- Run this in production where auto-seed is disabled.
-- Idempotent — safe to re-run.
-- ============================================================

-- 1. New role
INSERT INTO roles (name) VALUES ('sales')
ON CONFLICT (name) DO NOTHING;

-- 2. New permissions
INSERT INTO permissions (name) VALUES
    ('sales:manage-referral'),
    ('commission:read_own'),
    ('commission:read_all'),
    ('commission:manage_rate')
ON CONFLICT (name) DO NOTHING;

-- 3. Role → permission mappings

-- sales: manage own referral code + view own commission history
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'sales'
  AND p.name IN ('dashboard.view', 'session.read', 'sales:manage-referral', 'commission:read_own')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- superadmin: full commission reporting + rate control
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND p.name IN ('commission:read_all', 'commission:manage_rate')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- admin: full commission reporting + rate control
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN ('commission:read_all', 'commission:manage_rate')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );
