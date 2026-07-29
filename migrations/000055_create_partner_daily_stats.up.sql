CREATE TABLE IF NOT EXISTS partner_daily_stats (
    id SERIAL PRIMARY KEY,
    partner_external_id TEXT NOT NULL,
    date DATE NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(partner_external_id, date)
);

CREATE INDEX IF NOT EXISTS idx_partner_daily_stats_partner ON partner_daily_stats (partner_external_id);
CREATE INDEX IF NOT EXISTS idx_partner_daily_stats_date ON partner_daily_stats (date);
