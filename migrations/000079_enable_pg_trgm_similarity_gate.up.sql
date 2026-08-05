-- pg_trgm extension must be installed by a superuser before this migration runs.
-- Run once manually: CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- Then restart the server — this migration only creates the index.

CREATE INDEX IF NOT EXISTS idx_destinations_desc_trgm
  ON destinations USING GIST (description gist_trgm_ops)
  WHERE content_status = 'published';
