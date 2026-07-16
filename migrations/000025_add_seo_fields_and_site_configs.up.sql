ALTER TABLE destinations ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255) DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS seo_keywords TEXT DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS seo_description TEXT DEFAULT '';
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS og_image_url VARCHAR(500) DEFAULT '';

CREATE TABLE site_configs (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT DEFAULT '',
    category VARCHAR(50) DEFAULT 'general',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_site_configs_key ON site_configs(key);
CREATE INDEX idx_site_configs_category ON site_configs(category);

INSERT INTO site_configs (key, value, category) VALUES
('site_title', 'Jogjagem — Jelajahi Yogyakarta dengan AI', 'seo'),
('site_description', 'Temukan destinasi wisata terbaik di Yogyakarta dengan rekomendasi AI. Panduan lengkap Candi Prambanan, Malioboro, Pantai Parangtritis, Gunung Merapi, dan 100+ destinasi lainnya.', 'seo'),
('site_keywords', 'wisata Yogyakarta, jogja, travel guide Yogyakarta, destinasi wisata jogja, Candi Prambanan, Malioboro, Pantai Parangtritis, Gunung Merapi, AI tourism, paket wisata jogja', 'seo'),
('og_default_image', '/og-default.png', 'seo'),
('twitter_handle', '@jogjagem', 'seo'),
('landing_hero_title', 'Jogja is\nCalling You', 'landing'),
('landing_hero_subtitle', 'Discover unforgettable places, authentic experiences, and warm Javanese hospitality.', 'landing'),
('landing_cta_text', 'Mulai Jelajahi', 'landing')
ON CONFLICT (key) DO NOTHING;
