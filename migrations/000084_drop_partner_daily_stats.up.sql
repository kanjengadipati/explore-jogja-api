-- Phase 8: Drop the last legacy partners-related table.
-- partner_daily_stats was only written/read by the retired partner module.
-- tables partners and partner_applications were already dropped in 000069.

DROP TABLE IF EXISTS partner_daily_stats CASCADE;
