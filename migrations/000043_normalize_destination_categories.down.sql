-- 000043: Reverse category normalization
-- NOTE: This is best-effort. The original "Desa Wisata" stubs cannot be recovered
-- from this migration alone; they must be re-scraped from jadesta if needed.
-- Category labels are restored to their pre-migration values.

-- Undo SPLIT: cultural experiences -> "Cultural"
-- Undo MERGE: heritage/nature back to original labels
-- We reverse by checking source where available, but for simplicity we'll
-- only reverse the known direct merges.

-- Reverse "heritage" -> "Wisata" for known Borobudur/Ramu Boko/TMII entries
-- that originally had category = "Wisata" before migration.
UPDATE destinations SET category = 'Wisata'
  WHERE source IN ('manual', 'seed')
    AND category = 'heritage'
    AND (external_id IN ('borobudur', 'manohara', 'ramayana', 'ratu-boko', 'taman-mini-indonesia-indah'));

-- Reverse "nature" -> "Wisata Bukit" for Obelix entries
UPDATE destinations SET category = 'Wisata Bukit'
  WHERE external_id IN ('obelix-hills', 'obelix-sea-view')
    AND category = 'nature';

-- Reverse injourney "camping" -> "Cultural"
UPDATE destinations SET category = 'Cultural'
  WHERE source = 'injourney' AND category = 'camping';

-- Reverse injourney "culinary" -> "Cultural"
UPDATE destinations SET category = 'Cultural'
  WHERE source = 'injourney' AND category = 'culinary';

-- Reverse injourney "adventure" -> "Cultural"
UPDATE destinations SET category = 'Cultural'
  WHERE source = 'injourney' AND category = 'adventure';

-- Reverse injourney "family" -> "Cultural"
UPDATE destinations SET category = 'Cultural'
  WHERE source = 'injourney' AND category = 'family';

-- Reverse injourney "nature" -> "Cultural"
UPDATE destinations SET category = 'Cultural'
  WHERE source = 'injourney' AND category = 'nature';
