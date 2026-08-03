-- Phase 0 correction: change businesses.status default from 'approved' to 'pending'.
-- New self-service businesses should always start as pending (matching the
-- "Menunggu Verifikasi" dashboard flow). Backfill rows that exist from
-- seed_partner_business_backfill.sql already have explicit status values set,
-- so this only changes what future INSERT statements without explicit status get.
ALTER TABLE businesses ALTER COLUMN status SET DEFAULT 'pending';
