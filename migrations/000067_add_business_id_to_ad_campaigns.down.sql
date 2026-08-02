-- Phase 4 reversal: drop the business FK. partner_name is still present (it was
-- never dropped by the up migration — that is a manual post-verification step).
ALTER TABLE ad_campaigns DROP COLUMN IF EXISTS business_external_id;
