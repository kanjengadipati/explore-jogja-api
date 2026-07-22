-- Backfill empty sub_region.

-- InJourney / Borobudur / Ratu Boko area → Sleman
UPDATE destinations SET sub_region = 'Sleman'
WHERE (sub_region = '' OR sub_region IS NULL)
  AND id IN (
    'borobudur', 'ratu-boko', 'ramayana', 'manohara', 'taman-mini-indonesia-indah',
    'borobudur-sunset-meditation-class', 'waisak-di-borobudur', 'mahakarya-borobudur', 'borobudur-sunset',
    'paket-boko-picnic', 'ratu-boko-sunset', 'boko-membatik', 'boko-wedding', 'boko-prewedding',
    'boko-camping', 'boko-racik-rimpang', 'dhaharan-bandung-bondowoso', 'andrawina-barbekyu', 'boko-trekking'
  );

-- Remaining (Jadesta Desa Wisata) → Yogyakarta (default, corrected on next scrape)
UPDATE destinations SET sub_region = 'Yogyakarta'
WHERE sub_region = '' OR sub_region IS NULL;
