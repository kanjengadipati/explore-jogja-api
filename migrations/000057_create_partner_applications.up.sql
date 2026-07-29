CREATE TABLE IF NOT EXISTS partner_applications (
    id SERIAL PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    applicant_user_id INTEGER NOT NULL,

    business_name TEXT NOT NULL,
    category TEXT NOT NULL,
    location TEXT,
    phone TEXT,

    status TEXT NOT NULL DEFAULT 'pending',
    rejection_reason TEXT,

    converted_partner_external_id TEXT,

    submitted_at TIMESTAMP WITH TIME ZONE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewed_by INTEGER,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_partner_applications_status ON partner_applications (status);
CREATE INDEX IF NOT EXISTS idx_partner_applications_applicant ON partner_applications (applicant_user_id);
CREATE INDEX IF NOT EXISTS idx_partner_applications_deleted_at ON partner_applications (deleted_at);
