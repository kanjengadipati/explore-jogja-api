-- Revert site_title fix
UPDATE site_configs
SET value = 'Jogjagem — Jogja /n is', updated_at = NOW()
WHERE key = 'site_title' AND value = 'Jogjagem — Jelajahi Yogyakarta Lebih Dalam';
