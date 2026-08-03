CREATE TABLE IF NOT EXISTS listing_claims (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    business_id INTEGER NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    listing_type VARCHAR(50) NOT NULL,   -- 'destination' | 'hotel' | 'restaurant' | 'souvenir' | 'rental' | 'guide' | 'event'
    listing_external_id VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,
    submitted_at TIMESTAMP DEFAULT NOW(),
    reviewed_at TIMESTAMP,
    reviewed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

ALTER TABLE listing_claims DROP CONSTRAINT IF EXISTS chk_listing_claims_status;
ALTER TABLE listing_claims ADD CONSTRAINT chk_listing_claims_status
    CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE listing_claims DROP CONSTRAINT IF EXISTS chk_listing_claims_type;
ALTER TABLE listing_claims ADD CONSTRAINT chk_listing_claims_type
    CHECK (listing_type IN ('destination','hotel','restaurant','souvenir','rental','guide','event'));

CREATE INDEX IF NOT EXISTS idx_listing_claims_status ON listing_claims(status);
CREATE INDEX IF NOT EXISTS idx_listing_claims_business_id ON listing_claims(business_id);
-- Only one *approved* claim per listing should ever exist — enforced partially here,
-- fully enforced in the application layer on approval (see below).
CREATE INDEX IF NOT EXISTS idx_listing_claims_listing ON listing_claims(listing_type, listing_external_id);
