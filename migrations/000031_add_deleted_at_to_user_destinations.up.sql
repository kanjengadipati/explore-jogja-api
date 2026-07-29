-- deleted_at is now included in migration 000028. This migration is a safe no-op.
-- It adds the column only if the table exists AND the column is missing
-- (handles databases that ran the original 028 without deleted_at).
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'user_destinations'
  ) THEN
    IF NOT EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = 'public'
        AND table_name   = 'user_destinations'
        AND column_name  = 'deleted_at'
    ) THEN
      ALTER TABLE user_destinations ADD COLUMN deleted_at TIMESTAMP;
      CREATE INDEX idx_user_destinations_deleted_at ON user_destinations(deleted_at);
    END IF;
  END IF;
END $$;
