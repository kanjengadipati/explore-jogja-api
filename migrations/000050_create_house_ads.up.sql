CREATE TABLE IF NOT EXISTS house_ads (
    id SERIAL PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    placement TEXT NOT NULL UNIQUE,
    headline TEXT NOT NULL,
    subline TEXT,
    cta_label TEXT NOT NULL,
    image_url TEXT,
    target_url TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_house_ads_placement_enabled ON house_ads (placement, is_enabled);
CREATE INDEX IF NOT EXISTS idx_house_ads_deleted_at ON house_ads (deleted_at);
