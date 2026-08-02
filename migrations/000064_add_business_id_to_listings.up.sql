-- Phase 2: add business ownership to the 6 unclaimed listing types + events.
-- business_id is nullable; SET NULL on business delete so listings are not lost.
ALTER TABLE destinations ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE hotels       ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE restaurants  ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE souvenirs    ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE rentals      ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE guides       ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;
ALTER TABLE events       ADD COLUMN IF NOT EXISTS business_id INTEGER REFERENCES businesses(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_destinations_business_id ON destinations(business_id);
CREATE INDEX IF NOT EXISTS idx_hotels_business_id        ON hotels(business_id);
CREATE INDEX IF NOT EXISTS idx_restaurants_business_id   ON restaurants(business_id);
CREATE INDEX IF NOT EXISTS idx_souvenirs_business_id     ON souvenirs(business_id);
CREATE INDEX IF NOT EXISTS idx_rentals_business_id       ON rentals(business_id);
CREATE INDEX IF NOT EXISTS idx_guides_business_id        ON guides(business_id);
CREATE INDEX IF NOT EXISTS idx_events_business_id        ON events(business_id);
