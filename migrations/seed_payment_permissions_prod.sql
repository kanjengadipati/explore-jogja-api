-- ============================================================
-- Manual seed: payment permissions
-- Run this in production where auto-seed is disabled.
-- Idempotent — safe to re-run (uses ON CONFLICT DO NOTHING).
-- ============================================================

-- 1. New permissions
INSERT INTO permissions (name) VALUES
    ('payment.manage')
ON CONFLICT (name) DO NOTHING;

-- 2. Role → permission mappings
-- Helper: only insert if the pair doesn't already exist.

-- superadmin: add payment management permissions
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND p.name IN ('payment.manage')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- admin: add payment management permissions
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN ('payment.manage')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );
