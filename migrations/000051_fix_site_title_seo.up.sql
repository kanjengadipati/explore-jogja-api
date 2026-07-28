-- Fix site_title that was incorrectly set to 'Jogjagem — Jogja /n is'
UPDATE site_configs
SET value = 'Jogjagem — Jelajahi Yogyakarta Lebih Dalam', updated_at = NOW()
WHERE key = 'site_title' AND value = 'Jogjagem — Jogja /n is';
