-- Revert: set content_status back to '' for rows that were backfilled.
-- Only reverts rows that still have status='published' and were clearly backfilled
-- (no template_variant set means they never went through the AI pipeline).
UPDATE destinations
SET content_status = ''
WHERE status = 'published'
  AND content_status = 'published'
  AND template_variant = '';
