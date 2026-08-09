DROP INDEX IF EXISTS idx_destinations_hidden_gem_override;
ALTER TABLE destinations DROP COLUMN hidden_gem_pinned_at;
ALTER TABLE destinations DROP COLUMN hidden_gem_override;
