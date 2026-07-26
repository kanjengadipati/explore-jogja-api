-- 000043: Normalize destination categories
-- 1. PURGE  29 "Desa Wisata" stubs (jadesta scrape, not actually in DIY, empty content)
-- 2. MERGE  "Wisata"        -> "heritage"
-- 3. MERGE  "Wisata Bukit"  -> "nature"
-- 4. SPLIT  "Cultural" (injourney hardcode) -> proper categories
-- 5. CLEAN  strip stray HTML from descriptions

-- 1. PURGE Desa Wisata
DELETE FROM destinations WHERE category = 'Desa Wisata';

-- 2. MERGE Wisata -> heritage
UPDATE destinations SET category = 'heritage'
  WHERE category = 'Wisata';

-- 3. MERGE Wisata Bukit -> nature
UPDATE destinations SET category = 'nature'
  WHERE category = 'Wisata Bukit';

-- 4. SPLIT Cultural (injourney experiences) into proper categories
--    culinar y: BBQ / meals / picnic / racik
UPDATE destinations SET category = 'culinary'
  WHERE category = 'Cultural'
    AND (LOWER(name) LIKE '%barbekyu%' OR LOWER(name) LIKE '%bbq%'
      OR LOWER(name) LIKE '%meals%' OR LOWER(name) LIKE '%dhaharan%'
      OR LOWER(name) LIKE '%racik%' OR LOWER(name) LIKE '%picnic%');

--    camping
UPDATE destinations SET category = 'camping'
  WHERE category = 'Cultural'
    AND LOWER(name) LIKE '%camping%';

--    adventure: cycling / trekking / outbound
UPDATE destinations SET category = 'adventure'
  WHERE category = 'Cultural'
    AND (LOWER(name) LIKE '%cycling%' OR LOWER(name) LIKE '%trekking%'
      OR LOWER(name) LIKE '%outbound%');

--    family: sinema / cinema
UPDATE destinations SET category = 'family'
  WHERE category = 'Cultural'
    AND (LOWER(name) LIKE '%sinema%' OR LOWER(name) LIKE '%cinema%');

--    nature: bukit dagi / dagi abhinaya
UPDATE destinations SET category = 'nature'
  WHERE category = 'Cultural'
    AND LOWER(name) LIKE '%dagi%';

--    everything still Cultural -> heritage (default: all injourney
--    experiences are add-ons at heritage sites)
UPDATE destinations SET category = 'heritage'
  WHERE category = 'Cultural';

-- 5. CLEAN stray HTML tags from description
UPDATE destinations
  SET description = REGEXP_REPLACE(description, '<[^>]+>', ' ', 'g')
  WHERE description LIKE '%<%';

UPDATE destinations
  SET description = REGEXP_REPLACE(description, '\s+', ' ', 'g')
  WHERE description LIKE '%  %';

UPDATE destinations
  SET description = TRIM(description);
