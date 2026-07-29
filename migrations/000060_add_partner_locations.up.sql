ALTER TABLE partner_applications ADD COLUMN IF NOT EXISTS locations JSONB DEFAULT '[]'::jsonb;
