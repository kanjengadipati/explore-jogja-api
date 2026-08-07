-- migration: 000083_add_ecosystem_to_ad_campaigns
--
-- Ecosystem rail ("Rekomendasi Kebutuhan Traveler") placements.
--
-- A campaign on an ecosystem_* placement promotes a specific listing owned by
-- a business (listing_type + listing_external_id) and may target optional
-- destinations (target_dest_ids, JSONB array of destination external IDs —
-- empty array means "all destinations"). sort_order controls the card ranking
-- inside the rail (lower = higher), replacing the retired partner sponsor_tier.
--
-- This is additive: /partners/sponsored keeps serving until every frontend
-- caller is migrated, then it is deleted with the partner module.

ALTER TABLE ad_campaigns
    ADD COLUMN IF NOT EXISTS listing_type         VARCHAR(50),
    ADD COLUMN IF NOT EXISTS listing_external_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS target_dest_ids      JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS sort_order           INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_ad_campaigns_ecosystem_serving
    ON ad_campaigns(placement, is_active, payment_status, sort_order);
