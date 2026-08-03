ALTER TABLE destinations DROP COLUMN IF EXISTS business_id;
ALTER TABLE hotels       DROP COLUMN IF EXISTS business_id;
ALTER TABLE restaurants  DROP COLUMN IF EXISTS business_id;
ALTER TABLE souvenirs    DROP COLUMN IF EXISTS business_id;
ALTER TABLE rentals      DROP COLUMN IF EXISTS business_id;
ALTER TABLE guides       DROP COLUMN IF EXISTS business_id;
ALTER TABLE events       DROP COLUMN IF EXISTS business_id;
