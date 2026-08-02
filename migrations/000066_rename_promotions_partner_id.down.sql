-- Phase 4 reversal: drop the business FK, restore the partner_id column name.
ALTER TABLE promotions DROP COLUMN IF EXISTS business_external_id;
ALTER TABLE promotions RENAME COLUMN legacy_partner_external_id TO partner_id;
