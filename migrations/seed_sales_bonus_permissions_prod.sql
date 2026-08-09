-- ============================================================
-- Manual seed: sales bonus permissions
-- Run this in production where auto-seed is disabled.
-- Idempotent — safe to re-run.
-- ============================================================

-- 1. New permissions
INSERT INTO permissions (name) VALUES
    ('bonus:read_own'),
    ('bonus:read_all'),
    ('bonus:manage_rules'),
    ('bonus:manage_payout')
ON CONFLICT (name) DO NOTHING;

-- 2. Role → permission mappings

-- sales: view own bonus history
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'sales'
  AND p.name IN ('bonus:read_own')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- superadmin: full bonus reporting + rule management + payout marking
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND p.name IN ('bonus:read_all', 'bonus:manage_rules', 'bonus:manage_payout')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- admin: full bonus reporting + rule management + payout marking
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN ('bonus:read_all', 'bonus:manage_rules', 'bonus:manage_payout')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );
