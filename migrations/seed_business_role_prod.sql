-- ============================================================
-- Manual seed: business + listing_claim roles and permissions
-- Run this in production where auto-seed is disabled.
-- Idempotent — safe to re-run (uses ON CONFLICT DO NOTHING
-- and NOT EXISTS guards for role_permissions).
-- ============================================================

-- 1. New permissions
INSERT INTO permissions (name) VALUES
    ('business.read_own'),
    ('business.create_own'),
    ('business.update_own'),
    ('business.delete_own'),
    ('business.read_all'),
    ('business.approve'),
    ('business.reject'),
    ('business.suspend'),
    ('business.manage_all'),
    ('listing_claim.submit_own'),
    ('listing_claim.read_own'),
    ('listing_claim.read_all'),
    ('listing_claim.approve'),
    ('listing_claim.reject')
ON CONFLICT (name) DO NOTHING;

-- 2. superadmin: all business.* and listing_claim.* permissions.
--    Parenthesised so LIKE precedence does not leak listing_claim.* to all roles.
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'superadmin'
  AND (p.name LIKE 'business.%' OR p.name LIKE 'listing_claim.%')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- 3. admin: business management + listing claim review
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN (
      'business.read_all', 'business.approve', 'business.reject', 'business.suspend', 'business.manage_all',
      'listing_claim.read_all', 'listing_claim.approve', 'listing_claim.reject'
  )
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- 4. partner: business self-service + listing claim self-service (during transition)
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'partner'
  AND p.name IN (
      'business.read_own', 'business.create_own', 'business.update_own', 'business.delete_own',
      'listing_claim.submit_own', 'listing_claim.read_own'
  )
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );
