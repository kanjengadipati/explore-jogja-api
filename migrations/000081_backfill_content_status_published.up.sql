-- migration: 000081_backfill_content_status_published
-- Backfill content_status = 'published' for all existing destinations that:
--   1. Are already publicly visible (status = 'published'), AND
--   2. Have a non-empty description (i.e. have real content worth protecting from duplicate detection)
--   3. Were created before the content pipeline existed (content_status = '')
--
-- This ensures the pg_trgm similarity gate (migration 000079) compares new AI drafts
-- against the full existing catalog, not just destinations that have gone through the
-- new pipeline. Without this, ~375 existing destinations are invisible to similarity checks.

UPDATE destinations
SET content_status = 'published'
WHERE status = 'published'
  AND content_status = ''
  AND description != '';
