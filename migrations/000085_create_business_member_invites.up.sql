-- Business team-member invitations.
-- Lets an owner invite a person who may not have a Jogjagem account yet. The
-- invite carries a random token; the recipient registers/logs in and accepts
-- via the web portal, after which a business_owners row is created.
CREATE TABLE IF NOT EXISTS business_member_invites (
    id SERIAL PRIMARY KEY,
    business_id INTEGER NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    token_hash VARCHAR(64) NOT NULL,
    invited_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP NOT NULL,
    accepted_at TIMESTAMP,
    accepted_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_business_member_invites_business ON business_member_invites(business_id);
CREATE INDEX IF NOT EXISTS idx_business_member_invites_email ON business_member_invites(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_business_member_invites_token_hash ON business_member_invites(token_hash);
