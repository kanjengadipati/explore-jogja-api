-- Seed self-promo House Ads with CTA "Pasang Iklan".
-- Canonical reference for internal/seeds/seed_house_ads.go (SeedHouseAds).
-- Note: this file is intentionally not picked up by golang-migrate (no
-- numeric prefix), it exists as documentation only — the Go seed above is
-- what actually runs via `go run cmd/seed` (or AUTO_RUN_SEEDS).
INSERT INTO house_ads (external_id, placement, headline, headline_en, subline, subline_en, cta_label, cta_label_en, image_url, target_url, is_enabled)
VALUES
  ('house_ad_promo_hero', 'homepage_hero_aicard',
   'Ribuan wisatawan lihat halaman ini tiap hari',
   'Thousands of travelers see this page every day',
   'Pasang bisnismu di posisi paling depan Jogjagem.',
   'Put your business at the very front of Jogjagem.',
   'Pasang Iklan', 'Advertise Now',
   '/images/house-ads/promo-hero.jpg',
   '/ads?placement=homepage_hero_aicard', true),

  ('house_ad_promo_listing_top', 'listing_top',
   'Bisnismu bisa tampil paling atas di sini',
   'Your business can appear at the very top here',
   'Slot ini kosong sekarang — jadi yang pertama dilihat pengunjung.',
   'This slot is empty right now — be the first thing visitors see.',
   'Pasang Iklan', 'Advertise Now',
   '/images/house-ads/promo-listing-top.jpg',
   '/ads?placement=listing_top', true),

  ('house_ad_promo_listing_native', 'listing_native',
   'Muncul natural di antara listing populer',
   'Show up naturally among popular listings',
   'Terintegrasi langsung di alur pencarian pengunjung.',
   'Integrated right into the visitor''s search flow.',
   'Pasang Iklan', 'Advertise Now',
   '/images/house-ads/promo-listing-native.jpg',
   '/ads?placement=listing_native', true),

  ('house_ad_promo_destination', 'destination_detail',
   'Promosikan usahamu ke pengunjung destinasi ini',
   'Promote your business to visitors of this destination',
   'Tampil tepat saat wisatawan sedang merencanakan kunjungan.',
   'Shows exactly when travelers are planning their visit.',
   'Pasang Iklan', 'Advertise Now',
   '/images/house-ads/promo-destination.jpg',
   '/ads?placement=destination_detail', true),

  ('house_ad_promo_trending', 'homepage_hero_trending',
   'Trending destinasi pilihan wisatawan',
   'Trending destinations chosen by travelers',
   'Pasang bisnismu di sini untuk menjangkau pencari inspirasi liburan.',
   'Place your business here to reach holiday inspiration seekers.',
   'Pasang Iklan', 'Advertise Now',
   '/images/house-ads/promo-trending.jpg',
   '/ads?placement=homepage_hero_trending', true)
-- Upsert: only repair rows that are empty/stale (e.g. admin-created placeholder rows
-- with blank headline). Never overwrite placements that already have real content.
ON CONFLICT (placement) DO UPDATE SET
  external_id = EXCLUDED.external_id,
  headline = EXCLUDED.headline,
  headline_en = EXCLUDED.headline_en,
  subline = EXCLUDED.subline,
  subline_en = EXCLUDED.subline_en,
  cta_label = EXCLUDED.cta_label,
  cta_label_en = EXCLUDED.cta_label_en,
  image_url = EXCLUDED.image_url,
  target_url = EXCLUDED.target_url,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW()
WHERE house_ads.headline IS NULL OR house_ads.headline = '';
