-- migration: 000083_add_ecosystem_to_ad_campaigns (down)

DROP INDEX IF EXISTS idx_ad_campaigns_ecosystem_serving;

ALTER TABLE ad_campaigns
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS target_dest_ids,
    DROP COLUMN IF EXISTS listing_external_id,
    DROP COLUMN IF EXISTS listing_type;
