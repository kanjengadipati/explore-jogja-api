-- Redo of migration 000072_add_address_and_service_areas_to_businesses.
-- The original file was named without the required ".up.sql" suffix, so
-- golang-migrate silently skipped it and prod never received the address
-- column or the business_service_areas table. IF NOT EXISTS makes this
-- safe to run even if already applied manually.
ALTER TABLE businesses ADD COLUMN IF NOT EXISTS address VARCHAR(500);

CREATE TABLE IF NOT EXISTS business_service_areas (
    id         BIGSERIAL PRIMARY KEY,
    business_id BIGINT NOT NULL REFERENCES businesses(id) ON DELETE CASCADE,
    region      VARCHAR(100) NOT NULL,
    UNIQUE(business_id, region)
);
