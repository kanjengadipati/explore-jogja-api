-- ============================================================
-- Sales bonus/commission scheme — manual seed for production
-- (auto-seed disabled). Safe to re-run.
-- ============================================================

-- 1. Role
INSERT INTO roles (name) VALUES ('sales')
ON CONFLICT (name) DO NOTHING;

-- 2. Permissions
INSERT INTO permissions (name) VALUES
    ('sales:manage-referral'),
    ('commission:read_own'),
    ('commission:read_all'),
    ('commission:manage_rate'),
    ('bonus:read_own'),
    ('bonus:read_all'),
    ('bonus:manage_rules'),
    ('bonus:manage_payout')
ON CONFLICT (name) DO NOTHING;

-- 3. Role → permission mappings
INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name = 'sales'
  AND p.name IN ('dashboard.view', 'session.read', 'sales:manage-referral', 'commission:read_own', 'bonus:read_own')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

INSERT INTO role_permissions (role_id, permission)
SELECT r.id, p.name
FROM roles r, permissions p
WHERE r.name IN ('superadmin', 'admin')
  AND p.name IN ('commission:read_all', 'commission:manage_rate', 'bonus:read_all', 'bonus:manage_rules', 'bonus:manage_payout')
  AND NOT EXISTS (
      SELECT 1 FROM role_permissions rp
      WHERE rp.role_id = r.id AND rp.permission = p.name AND rp.deleted_at IS NULL
  );

-- 4. Commission rate (optional — code falls back to 0.20)
INSERT INTO site_configs (key, value, category, created_at, updated_at)
VALUES ('sales_commission_rate', '0.20', 'sales', now(), now())
ON CONFLICT (key) DO NOTHING;

-- 5. Bonus rules — WITHOUT these, no bonuses are ever recorded.
--    Adjust amounts/thresholds to match your scheme.
INSERT INTO bonus_rules (type, metric, tier, threshold, amount, is_active)
SELECT v.type, v.metric, v.tier, v.threshold, v.amount, v.is_active
FROM (VALUES
    ('onboarding',    'tenant',      NULL::int, NULL::int, 50000::float8, TRUE),
    ('milestone',     'tenant',      1,         3,         25000,         TRUE),
    ('milestone',     'tenant',      2,         10,        75000,         TRUE),
    ('milestone',     'transaction', 1,         5,         10000,         TRUE),
    ('milestone',     'transaction', 2,         10,        35000,         FALSE)
) AS v(type, metric, tier, threshold, amount, is_active)
WHERE NOT EXISTS (
    SELECT 1 FROM bonus_rules r
    WHERE r.type = v.type AND r.metric = v.metric
      AND r.tier IS NOT DISTINCT FROM v.tier
      AND r.threshold IS NOT DISTINCT FROM v.threshold
      AND r.deleted_at IS NULL
);
