-- Phase 7: Drop legacy partners and partner_applications tables
-- This migration is applied after full cutover to businesses and listing_claims is verified.

DROP TABLE IF EXISTS partner_applications CASCADE;
DROP TABLE IF EXISTS partners CASCADE;
