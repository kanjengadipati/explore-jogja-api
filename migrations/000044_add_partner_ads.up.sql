ALTER TABLE partners
    ADD COLUMN IF NOT EXISTS is_sponsored BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS sponsor_tier INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS sponsor_start_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS sponsor_end_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS target_dest_ids JSONB DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS impression_count BIGINT DEFAULT 0,
    ADD COLUMN IF NOT EXISTS click_count BIGINT DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_partners_is_sponsored ON partners(is_sponsored);
CREATE INDEX IF NOT EXISTS idx_partners_target_dest_ids ON partners USING GIN(target_dest_ids);

CREATE TABLE IF NOT EXISTS ad_campaigns (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    partner_name VARCHAR(255) NOT NULL,
    placement VARCHAR(100) NOT NULL,
    image_url VARCHAR(1000) NOT NULL,
    target_url VARCHAR(1000) NOT NULL,
    category VARCHAR(100),
    start_at TIMESTAMP,
    end_at TIMESTAMP,
    weight INTEGER DEFAULT 1,
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ad_campaigns_external_id ON ad_campaigns(external_id);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_placement ON ad_campaigns(placement);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_category ON ad_campaigns(category);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_is_active ON ad_campaigns(is_active);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_deleted_at ON ad_campaigns(deleted_at);
