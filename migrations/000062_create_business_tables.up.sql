-- Business identity domain (Phase 0).
-- New tables for the Partner -> Business migration. Nothing reads from these
-- yet; they are scaffolded ahead of the later cutover phases.
CREATE TABLE IF NOT EXISTS businesses (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(500),
    avatar_url VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'approved',
    rejection_reason TEXT,
    submitted_at TIMESTAMP,
    reviewed_at TIMESTAMP,
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    legacy_partner_external_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_businesses_external_id ON businesses(external_id);
CREATE INDEX IF NOT EXISTS idx_businesses_status ON businesses(status);
CREATE INDEX IF NOT EXISTS idx_businesses_category ON businesses(category);
CREATE INDEX IF NOT EXISTS idx_businesses_deleted_at ON businesses(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_businesses_legacy_partner ON businesses(legacy_partner_external_id)
    WHERE legacy_partner_external_id IS NOT NULL;

-- Idempotent: a previous partial run may have already added the constraint,
-- and the dirty-migration retry path re-applies this file.
ALTER TABLE businesses DROP CONSTRAINT IF EXISTS chk_businesses_status;
ALTER TABLE businesses ADD CONSTRAINT chk_businesses_status
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'suspended'));

CREATE TABLE IF NOT EXISTS business_owners (
    id SERIAL PRIMARY KEY,
    business_id INTEGER NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_business_owners_unique ON business_owners(business_id, user_id);
CREATE INDEX IF NOT EXISTS idx_business_owners_user_id ON business_owners(user_id);
