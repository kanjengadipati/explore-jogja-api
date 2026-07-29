-- Add ownership and approval workflow columns to partners.
-- Default status='approved' keeps existing admin-seeded listings visible.

ALTER TABLE partners
    ADD COLUMN IF NOT EXISTS owner_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved',
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
    ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_partners_status ON partners(status);
CREATE INDEX IF NOT EXISTS idx_partners_owner_user_id ON partners(owner_user_id);

ALTER TABLE partners ADD CONSTRAINT chk_partners_status
    CHECK (status IN ('pending', 'approved', 'rejected', 'suspended'));
