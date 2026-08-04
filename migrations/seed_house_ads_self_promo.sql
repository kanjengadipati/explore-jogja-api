-- Seed self-promo House Ads with CTA "Pasang Iklan"
INSERT INTO house_ads (external_id, placement, headline, subline, cta_label, image_url, target_url, is_enabled)
VALUES
  ('house_ad_promo_hero', 'homepage_hero_aicard',
   'Ribuan wisatawan lihat halaman ini tiap hari',
   'Pasang bisnismu di posisi paling depan Jogjagem.',
   'Pasang Iklan', '/images/house-ads/promo-hero.jpg',
   '/ads?placement=homepage_hero_aicard', true),

  ('house_ad_promo_listing_top', 'listing_top',
   'Bisnismu bisa tampil paling atas di sini',
   'Slot ini kosong sekarang — jadi yang pertama dilihat pengunjung.',
   'Pasang Iklan', '/images/house-ads/promo-listing-top.jpg',
   '/ads?placement=listing_top', true),

  ('house_ad_promo_listing_native', 'listing_native',
   'Muncul natural di antara listing populer',
   'Terintegrasi langsung di alur pencarian pengunjung.',
   'Pasang Iklan', '/images/house-ads/promo-listing-native.jpg',
   '/ads?placement=listing_native', true),

  ('house_ad_promo_destination', 'destination_detail',
   'Promosikan usahamu ke pengunjung destinasi ini',
   'Tampil tepat saat wisatawan sedang merencanakan kunjungan.',
   'Pasang Iklan', '/images/house-ads/promo-destination.jpg',
   '/ads?placement=destination_detail', true)
-- Upsert: only repair rows that are empty/stale (e.g. admin-created placeholder rows
-- with blank headline). Never overwrite placements that already have real content.
ON CONFLICT (placement) DO UPDATE SET
  external_id = EXCLUDED.external_id,
  headline = EXCLUDED.headline,
  subline = EXCLUDED.subline,
  cta_label = EXCLUDED.cta_label,
  image_url = EXCLUDED.image_url,
  target_url = EXCLUDED.target_url,
  is_enabled = EXCLUDED.is_enabled,
  updated_at = NOW()
WHERE house_ads.headline IS NULL OR house_ads.headline = '';
