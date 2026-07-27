-- Scope reviews to partner listings and allow partner reply.
-- partner_id is nullable — pre-Phase-4 reviews (destination-only) are unaffected.

ALTER TABLE reviews
    ADD COLUMN IF NOT EXISTS partner_id VARCHAR(100) REFERENCES partners(external_id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reply TEXT,
    ADD COLUMN IF NOT EXISTS replied_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS replied_by INTEGER REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_reviews_partner_id ON reviews(partner_id);
