-- migration: 000072_add_address_and_service_areas_to_businesses
-- Adds the address column (nullable — existing rows have no valid default) and
-- the business_service_areas table for multi-region service area support.
-- Run this before deploying backend code that writes to these fields.

-- up

ALTER TABLE businesses ADD COLUMN IF NOT EXISTS address VARCHAR(500);

CREATE TABLE IF NOT EXISTS business_service_areas (
    id         BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    region      VARCHAR(100) NOT NULL,
    UNIQUE(business_id, region)
);

-- down

DROP TABLE IF EXISTS business_service_areas;
ALTER TABLE businesses DROP COLUMN IF EXISTS address;
