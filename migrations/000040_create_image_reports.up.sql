CREATE TABLE IF NOT EXISTS image_reports (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    destination_id TEXT NOT NULL,
    image_url TEXT,
    user_id BIGINT,
    user_name TEXT,
    reason TEXT NOT NULL,
    details TEXT,
    status TEXT DEFAULT 'pending'
);

CREATE INDEX IF NOT EXISTS idx_image_reports_destination_id ON image_reports(destination_id);
CREATE INDEX IF NOT EXISTS idx_image_reports_status ON image_reports(status);
CREATE INDEX IF NOT EXISTS idx_image_reports_deleted_at ON image_reports(deleted_at);
