-- Reverse backfill (best effort — original empty values are lost)
UPDATE destinations SET sub_region = '' WHERE sub_region = 'Yogyakarta';
UPDATE destinations SET sub_region = '' WHERE sub_region = 'Sleman' AND id IN (
  'borobudur', 'ratu-boko', 'ramayana', 'manohara', 'taman-mini-indonesia-indah',
  'borobudur-sunset-meditation-class', 'waisak-di-borobudur', 'mahakarya-borobudur', 'borobudur-sunset',
  'paket-boko-picnic', 'ratu-boko-sunset', 'boko-membatik', 'boko-wedding', 'boko-prewedding',
  'boko-camping', 'boko-racik-rimpang', 'dhaharan-bandung-bondowoso', 'andrawina-barbekyu', 'boko-trekking'
);
