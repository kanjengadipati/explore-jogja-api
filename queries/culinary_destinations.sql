-- ============================================================
-- Query: 90 Destinasi Kuliner BARU (tambahan 26 Jul 2026)
-- 10 destinasi existing dikecualikan
-- ============================================================

-- 1. Semua kuliner baru — kolom penting
SELECT
    id,
    external_id,
    name,
    tagline,
    category,
    location,
    sub_region,
    rating,
    review_count,
    description,
    ticket_price,
    opening_hours,
    best_time,
    latitude,
    longitude,
    google_maps_url,
    seo_title,
    name_en,
    tagline_en,
    created_at
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id NOT IN (
    'malioboro',
    'kopi-klotok-pakem',
    'gudeg-wijilan',
    'pasar-beringharjo',
    'kasongan-keramik',
    'kopi-ampirono',
    'pasar-kaki-langit-mangunan',
    'kopi-ingkar-janji',
    'sentra-kerajinan-bambu-cebongan',
    'soto-kadipiro'
  )
ORDER BY sub_region, rating DESC, name;

-- 2. Ringkasan per sub_region (kuliner baru saja)
SELECT
    sub_region,
    COUNT(*) AS jumlah,
    ROUND(AVG(rating), 2) AS avg_rating,
    SUM(review_count) AS total_reviews
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id NOT IN (
    'malioboro','kopi-klotok-pakem','gudeg-wijilan','pasar-beringharjo',
    'kasongan-keramik','kopi-ampirono','pasar-kaki-langit-mangunan',
    'kopi-ingkar-janji','sentra-kerajinan-bambu-cebongan','soto-kadipiro'
  )
GROUP BY sub_region
ORDER BY jumlah DESC;

-- 3. Top 10 kuliner baru rating tertinggi
SELECT
    external_id,
    name,
    sub_region,
    rating,
    review_count,
    ticket_price
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id NOT IN (
    'malioboro','kopi-klotok-pakem','gudeg-wijilan','pasar-beringharjo',
    'kasongan-keramik','kopi-ampirono','pasar-kaki-langit-mangunan',
    'kopi-ingkar-janji','sentra-kerajinan-bambu-cebongan','soto-kadipiro'
  )
ORDER BY rating DESC, review_count DESC
LIMIT 10;

-- 4. Detail kuliner baru per sub_region
SELECT
    sub_region,
    name,
    rating,
    review_count,
    ticket_price,
    opening_hours
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id NOT IN (
    'malioboro','kopi-klotok-pakem','gudeg-wijilan','pasar-beringharjo',
    'kasongan-keramik','kopi-ampirono','pasar-kaki-langit-mangunan',
    'kopi-ingkar-janji','sentra-kerajinan-bambu-cebongan','soto-kadipiro'
  )
ORDER BY
    CASE sub_region
        WHEN 'Yogyakarta' THEN 1
        WHEN 'Sleman' THEN 2
        WHEN 'Bantul' THEN 3
        WHEN 'Gunungkidul' THEN 4
        WHEN 'Kulon Progo' THEN 5
        ELSE 6
    END,
    rating DESC;

-- 5. Full export kuliner baru (semua kolom termasuk JSONB)
SELECT
    external_id AS id,
    name,
    tagline,
    category,
    location,
    sub_region,
    images,
    rating,
    review_count,
    description,
    story,
    ticket_price,
    opening_hours,
    facilities,
    travel_tips,
    best_time,
    weather,
    latitude,
    longitude,
    reviews,
    partners,
    faqs,
    google_maps_url,
    google_review_count,
    seo_title,
    seo_keywords,
    seo_description,
    og_image_url,
    video_url,
    name_en,
    tagline_en,
    description_en,
    story_en,
    best_time_en,
    facilities_en,
    travel_tips_en,
    seo_title_en,
    seo_keywords_en,
    seo_description_en
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id NOT IN (
    'malioboro','kopi-klotok-pakem','gudeg-wijilan','pasar-beringharjo',
    'kasongan-keramik','kopi-ampirono','pasar-kaki-langit-mangunan',
    'kopi-ingkar-janji','sentra-kerajinan-bambu-cebongan','soto-kadipiro'
  )
  ORDER BY sub_region, name;

-- ============================================================
-- Query: 47 Destinasi Populer BARU (Semua Kategori, 26 Jul 2026)
-- ============================================================

-- 11. Semua destinasi populer baru — kolom penting
SELECT
    id,
    external_id,
    name,
    tagline,
    category,
    location,
    sub_region,
    rating,
    review_count,
    description,
    ticket_price,
    opening_hours,
    best_time,
    latitude,
    longitude,
    google_maps_url,
    seo_title,
    name_en,
    tagline_en,
    created_at
FROM destinations
WHERE deleted_at IS NULL
  AND external_id IN (
    'museum-perjuangan','kraton-kauman',
    'hutan-pinus-pengger','tlogo-putri-kaliurang','taman-pelangi-jogja',
    'terowongan-jembatan-api','pulesari','goa-panyingkisan','gondosuli',
    'sumber-air-tempaksari','situ-gede','pemandian-air-panas-cangar','puncak-gramat',
    'pantai-ngrune','pantai-kuwaru','pantai-singkahan','pantai-pandansimo','pantai-patehan',
    'river-tubing-sungai-epoh','atv-merapi','horse-riding-parangtritis','outbound-mungkid',
    'taman-rekreasi-kidzania','bale-cangar','kedung-pedut',
    'desa-budaya-prambanan','batik-kadipaten','kampung-batik-jogja',
    'museum-ranggawarsita','museum-telekomunikasi',
    'gado-gado-bopkri','nasi-gudeg-ambarrukmo','lesehan-gambir-anom','warung-sate-mbecak',
    'pindul-cave-tubing','timang-bridge',
    'goa-jomblang-cave','benteng-vredeburg-museum','pantai-sadranan-snorkeling',
    'ekowisata-mungkid','pantai-mesra-ngrawe','candi-ratu-boko-heritage',
    'heha-ocean-cliff','pantai-baron-wonosari',
    'candi-ikil','museum-biologi-ugm'
  )
ORDER BY category, sub_region, rating DESC, name;

-- 12. Ringkasan per kategori (populer baru saja)
SELECT
    category,
    COUNT(*) AS jumlah,
    ROUND(AVG(rating), 2) AS avg_rating,
    SUM(review_count) AS total_reviews
FROM destinations
WHERE deleted_at IS NULL
  AND external_id IN (
    'museum-perjuangan','kraton-kauman',
    'hutan-pinus-pengger','tlogo-putri-kaliurang','taman-pelangi-jogja',
    'terowongan-jembatan-api','pulesari','goa-panyingkisan','gondosuli',
    'sumber-air-tempaksari','situ-gede','pemandian-air-panas-cangar','puncak-gramat',
    'pantai-ngrune','pantai-kuwaru','pantai-singkahan','pantai-pandansimo','pantai-patehan',
    'river-tubing-sungai-epoh','atv-merapi','horse-riding-parangtritis','outbound-mungkid',
    'taman-rekreasi-kidzania','bale-cangar','kedung-pedut',
    'desa-budaya-prambanan','batik-kadipaten','kampung-batik-jogja',
    'museum-ranggawarsita','museum-telekomunikasi',
    'gado-gado-bopkri','nasi-gudeg-ambarrukmo','lesehan-gambir-anom','warung-sate-mbecak',
    'pindul-cave-tubing','timang-bridge',
    'goa-jomblang-cave','benteng-vredeburg-museum','pantai-sadranan-snorkeling',
    'ekowisata-mungkid','pantai-mesra-ngrawe','candi-ratu-boko-heritage',
    'heha-ocean-cliff','pantai-baron-wonosari',
    'candi-ikil','museum-biologi-ugm'
  )
GROUP BY category
ORDER BY jumlah DESC;

-- 13. Top 15 populer baru rating tertinggi
SELECT
    external_id,
    name,
    category,
    sub_region,
    rating,
    review_count,
    ticket_price
FROM destinations
WHERE deleted_at IS NULL
  AND external_id IN (
    'museum-perjuangan','kraton-kauman',
    'hutan-pinus-pengger','tlogo-putri-kaliurang','taman-pelangi-jogja',
    'terowongan-jembatan-api','pulesari','goa-panyingkisan','gondosuli',
    'sumber-air-tempaksari','situ-gede','pemandian-air-panas-cangar','puncak-gramat',
    'pantai-ngrune','pantai-kuwaru','pantai-singkahan','pantai-pandansimo','pantai-patehan',
    'river-tubing-sungai-epoh','atv-merapi','horse-riding-parangtritis','outbound-mungkid',
    'taman-rekreasi-kidzania','bale-cangar','kedung-pedut',
    'desa-budaya-prambanan','batik-kadipaten','kampung-batik-jogja',
    'museum-ranggawarsita','museum-telekomunikasi',
    'gado-gado-bopkri','nasi-gudeg-ambarrukmo','lesehan-gambir-anom','warung-sate-mbecak',
    'pindul-cave-tubing','timang-bridge',
    'goa-jomblang-cave','benteng-vredeburg-museum','pantai-sadranan-snorkeling',
    'ekowisata-mungkid','pantai-mesra-ngrawe','candi-ratu-boko-heritage',
    'heha-ocean-cliff','pantai-baron-wonosari',
    'candi-ikil','museum-biologi-ugm'
  )
ORDER BY rating DESC, review_count DESC
LIMIT 15;

-- 14. Detail populer baru per sub_region
SELECT
    sub_region,
    category,
    name,
    rating,
    review_count,
    ticket_price,
    opening_hours
FROM destinations
WHERE deleted_at IS NULL
  AND external_id IN (
    'museum-perjuangan','kraton-kauman',
    'hutan-pinus-pengger','tlogo-putri-kaliurang','taman-pelangi-jogja',
    'terowongan-jembatan-api','pulesari','goa-panyingkisan','gondosuli',
    'sumber-air-tempaksari','situ-gede','pemandian-air-panas-cangar','puncak-gramat',
    'pantai-ngrune','pantai-kuwaru','pantai-singkahan','pantai-pandansimo','pantai-patehan',
    'river-tubing-sungai-epoh','atv-merapi','horse-riding-parangtritis','outbound-mungkid',
    'taman-rekreasi-kidzania','bale-cangar','kedung-pedut',
    'desa-budaya-prambanan','batik-kadipaten','kampung-batik-jogja',
    'museum-ranggawarsita','museum-telekomunikasi',
    'gado-gado-bopkri','nasi-gudeg-ambarrukmo','lesehan-gambir-anom','warung-sate-mbecak',
    'pindul-cave-tubing','timang-bridge',
    'goa-jomblang-cave','benteng-vredeburg-museum','pantai-sadranan-snorkeling',
    'ekowisata-mungkid','pantai-mesra-ngrawe','candi-ratu-boko-heritage',
    'heha-ocean-cliff','pantai-baron-wonosari',
    'candi-ikil','museum-biologi-ugm'
  )
ORDER BY
    CASE sub_region
        WHEN 'Yogyakarta' THEN 1
        WHEN 'Sleman' THEN 2
        WHEN 'Bantul' THEN 3
        WHEN 'Gunungkidul' THEN 4
        WHEN 'Kulon Progo' THEN 5
        WHEN 'Magelang' THEN 6
        WHEN 'Semarang' THEN 7
        ELSE 8
    END,
    category, rating DESC;

-- 15. Full export populer baru (semua kolom termasuk JSONB)
SELECT
    external_id AS id,
    name,
    tagline,
    category,
    location,
    sub_region,
    images,
    rating,
    review_count,
    description,
    story,
    ticket_price,
    opening_hours,
    facilities,
    travel_tips,
    best_time,
    weather,
    latitude,
    longitude,
    reviews,
    partners,
    faqs,
    google_maps_url,
    google_review_count,
    seo_title,
    seo_keywords,
    seo_description,
    og_image_url,
    video_url,
    name_en,
    tagline_en,
    description_en,
    story_en,
    best_time_en,
    facilities_en,
    travel_tips_en,
    seo_title_en,
    seo_keywords_en,
    seo_description_en
FROM destinations
WHERE deleted_at IS NULL
  AND external_id IN (
    'museum-perjuangan','kraton-kauman',
    'hutan-pinus-pengger','tlogo-putri-kaliurang','taman-pelangi-jogja',
    'terowongan-jembatan-api','pulesari','goa-panyingkisan','gondosuli',
    'sumber-air-tempaksari','situ-gede','pemandian-air-panas-cangar','puncak-gramat',
    'pantai-ngrune','pantai-kuwaru','pantai-singkahan','pantai-pandansimo','pantai-patehan',
    'river-tubing-sungai-epoh','atv-merapi','horse-riding-parangtritis','outbound-mungkid',
    'taman-rekreasi-kidzania','bale-cangar','kedung-pedut',
    'desa-budaya-prambanan','batik-kadipaten','kampung-batik-jogja',
    'museum-ranggawarsita','museum-telekomunikasi',
    'gado-gado-bopkri','nasi-gudeg-ambarrukmo','lesehan-gambir-anom','warung-sate-mbecak',
    'pindul-cave-tubing','timang-bridge',
    'goa-jomblang-cave','benteng-vredeburg-museum','pantai-sadranan-snorkeling',
    'ekowisata-mungkid','pantai-mesra-ngrawe','candi-ratu-boko-heritage',
    'heha-ocean-cliff','pantai-baron-wonosari',
    'candi-ikil','museum-biologi-ugm'
  )
ORDER BY category, sub_region, name;

-- ============================================================
-- Query: 10 Tengkleng/Kambing BARU (tambahan 26 Jul 2026)
-- ============================================================

-- 6. Semua tengkleng/kambing baru — kolom penting
SELECT
    id,
    external_id,
    name,
    tagline,
    category,
    location,
    sub_region,
    rating,
    review_count,
    description,
    ticket_price,
    opening_hours,
    best_time,
    latitude,
    longitude,
    google_maps_url,
    seo_title,
    name_en,
    tagline_en,
    created_at
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id IN (
    'tengkleng-bhenjoyo',
    'tengkleng-pawon-umi-nunung',
    'tengkleng-kambing-86',
    'tengkleng-hohah',
    'tengkleng-solo-timoho',
    'tengkleng-laweyan',
    'soto-tengkleng-pak-marno',
    'tongseng-pak-kribo',
    'sate-kambing-mbah-so',
    'tengkleng-ponco'
  )
ORDER BY sub_region, rating DESC, name;

-- 7. Ringkasan tengkleng/kambing baru per sub_region
SELECT
    sub_region,
    COUNT(*) AS jumlah,
    ROUND(AVG(rating), 2) AS avg_rating,
    SUM(review_count) AS total_reviews
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id IN (
    'tengkleng-bhenjoyo','tengkleng-pawon-umi-nunung','tengkleng-kambing-86',
    'tengkleng-hohah','tengkleng-solo-timoho','tengkleng-laweyan',
    'soto-tengkleng-pak-marno','tongseng-pak-kribo','sate-kambing-mbah-so',
    'tengkleng-ponco'
  )
GROUP BY sub_region
ORDER BY jumlah DESC;

-- 8. Top tengkleng/kambing rating tertinggi
SELECT
    external_id,
    name,
    sub_region,
    rating,
    review_count,
    ticket_price
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id IN (
    'tengkleng-bhenjoyo','tengkleng-pawon-umi-nunung','tengkleng-kambing-86',
    'tengkleng-hohah','tengkleng-solo-timoho','tengkleng-laweyan',
    'soto-tengkleng-pak-marno','tongseng-pak-kribo','sate-kambing-mbah-so',
    'tengkleng-ponco'
  )
ORDER BY rating DESC, review_count DESC;

-- 9. Detail tengkleng/kambing baru per sub_region
SELECT
    sub_region,
    name,
    rating,
    review_count,
    ticket_price,
    opening_hours
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id IN (
    'tengkleng-bhenjoyo','tengkleng-pawon-umi-nunung','tengkleng-kambing-86',
    'tengkleng-hohah','tengkleng-solo-timoho','tengkleng-laweyan',
    'soto-tengkleng-pak-marno','tongseng-pak-kribo','sate-kambing-mbah-so',
    'tengkleng-ponco'
  )
ORDER BY
    CASE sub_region
        WHEN 'Yogyakarta' THEN 1
        WHEN 'Sleman' THEN 2
        WHEN 'Bantul' THEN 3
        WHEN 'Gunungkidul' THEN 4
        WHEN 'Kulon Progo' THEN 5
        ELSE 6
    END,
    rating DESC;

-- 10. Full export tengkleng/kambing baru (semua kolom termasuk JSONB)
SELECT
    external_id AS id,
    name,
    tagline,
    category,
    location,
    sub_region,
    images,
    rating,
    review_count,
    description,
    story,
    ticket_price,
    opening_hours,
    facilities,
    travel_tips,
    best_time,
    weather,
    latitude,
    longitude,
    reviews,
    partners,
    faqs,
    google_maps_url,
    google_review_count,
    seo_title,
    seo_keywords,
    seo_description,
    og_image_url,
    video_url,
    name_en,
    tagline_en,
    description_en,
    story_en,
    best_time_en,
    facilities_en,
    travel_tips_en,
    seo_title_en,
    seo_keywords_en,
    seo_description_en
FROM destinations
WHERE category = 'culinary'
  AND deleted_at IS NULL
  AND external_id IN (
    'tengkleng-bhenjoyo','tengkleng-pawon-umi-nunung','tengkleng-kambing-86',
    'tengkleng-hohah','tengkleng-solo-timoho','tengkleng-laweyan',
    'soto-tengkleng-pak-marno','tongseng-pak-kribo','sate-kambing-mbah-so',
    'tengkleng-ponco'
  )
ORDER BY sub_region, name;
