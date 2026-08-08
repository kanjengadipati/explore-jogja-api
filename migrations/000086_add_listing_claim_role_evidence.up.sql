-- Add optional claim metadata: who is claiming (owner vs representative)
-- and optional supporting evidence URL (uploaded proof of ownership).
ALTER TABLE listing_claims ADD COLUMN IF NOT EXISTS role VARCHAR(30) NOT NULL DEFAULT 'owner';
ALTER TABLE listing_claims ADD COLUMN IF NOT EXISTS evidence_url TEXT;
