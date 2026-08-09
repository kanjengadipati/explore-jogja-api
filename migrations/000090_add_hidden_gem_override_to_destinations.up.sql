ALTER TABLE destinations ADD COLUMN hidden_gem_override VARCHAR(10) NOT NULL DEFAULT '';
ALTER TABLE destinations ADD COLUMN hidden_gem_pinned_at TIMESTAMPTZ NULL;

CREATE INDEX idx_destinations_hidden_gem_override
    ON destinations(hidden_gem_override)
    WHERE hidden_gem_override != '';
