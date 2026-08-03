-- ============================================================
-- Manual seed: backfill businesses from existing partners (Phase 1)
-- Run this in production once (after migration 000062 is applied).
-- Idempotent — safe to re-run (uses ON CONFLICT DO NOTHING).
-- This is the one-time backfill for pre-existing partner rows; the
-- application-level dual-write hook covers all writes going forward.
-- ============================================================

-- 1. Mirror every partners row into businesses.
--    business.external_id = 'biz_' || partners.external_id
--    business.legacy_partner_external_id = partners.external_id (join key)
INSERT INTO businesses (
    external_id, name, description, category, phone, website,
    status, rejection_reason, submitted_at, reviewed_at, reviewed_by,
    legacy_partner_external_id, created_at, updated_at
)
SELECT
    'biz_' || p.external_id,
    p.name,
    p.description,
    p.category,
    p.phone,
    p.website,
    p.status,
    p.rejection_reason,
    p.submitted_at,
    p.reviewed_at,
    p.reviewed_by,
    p.external_id,
    p.created_at,
    NOW()
FROM partners p
ON CONFLICT DO NOTHING;

-- 2. Mirror owner_user_id into business_owners.
INSERT INTO business_owners (business_id, user_id, created_at)
SELECT b.id, p.owner_user_id, NOW()
FROM partners p
JOIN businesses b ON b.legacy_partner_external_id = p.external_id
WHERE p.owner_user_id IS NOT NULL
ON CONFLICT (business_id, user_id) DO NOTHING;

-- ============================================================
-- Acceptance checks (run after backfill; all should return 0 rows)
-- ============================================================

-- A. Parity: every non-deleted partner has exactly one mirrored business.
SELECT p.external_id
FROM partners p
LEFT JOIN businesses b ON b.legacy_partner_external_id = p.external_id
WHERE p.deleted_at IS NULL
  AND (b.id IS NULL OR b.deleted_at IS NOT NULL);

-- B. Ownership: every partner with an owner has that owner mirrored.
SELECT p.external_id, p.name
FROM partners p
JOIN businesses b ON b.legacy_partner_external_id = p.external_id
LEFT JOIN business_owners bo ON bo.business_id = b.id AND bo.user_id = p.owner_user_id
WHERE p.owner_user_id IS NOT NULL
  AND bo.id IS NULL;

-- C. Unowned partners (informational: these are candidates for claims in Phase 2).
SELECT p.external_id, p.name
FROM partners p
LEFT JOIN businesses b ON b.legacy_partner_external_id = p.external_id
WHERE p.deleted_at IS NULL
  AND p.owner_user_id IS NULL;
