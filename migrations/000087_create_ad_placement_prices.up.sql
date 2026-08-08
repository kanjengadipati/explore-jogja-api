-- migration: 000087_create_ad_placement_prices
--
-- Customizable ad placement pricing. Jogjagem's default monthly rates used to
-- live only in code (adcampaign/pricing.go MonthlyPrices map). This table makes
-- them editable from the admin portal (price management) so staff can adjust
-- rates and run promos/discounts per placement without a deploy.
--
-- The code map stays as the fallback for placements not yet in this table and
-- as the seed source; monthly_rate (not the code map) is the effective price
-- once a row exists.
--
-- promo_pct is a percentage discount (0 = no promo). When promo_pct > 0 and the
-- optional promo window (promo_start_at/promo_end_at) is empty or covers "now",
-- the promo is active and the effective monthly rate = monthly_rate × (1 − pct/100).

CREATE TABLE IF NOT EXISTS ad_placement_prices (
    placement      VARCHAR(50) PRIMARY KEY,
    monthly_rate   BIGINT      NOT NULL,
    currency       VARCHAR(8)  NOT NULL DEFAULT 'IDR',
    promo_pct      NUMERIC(5,2) NOT NULL DEFAULT 0,
    promo_label    VARCHAR(120) NOT NULL DEFAULT '',
    promo_start_at TIMESTAMPTZ,
    promo_end_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed with the current default rates so the table is authoritative from day
-- one and the admin UI has something to show/edit immediately.
INSERT INTO ad_placement_prices (placement, monthly_rate) VALUES
    ('homepage_hero_aicard',     300000),
    ('homepage_hero_trending',   250000),
    ('homepage_category_banner', 350000),
    ('listing_top',              250000),
    ('listing_native',           200000),
    ('destination_detail',       400000),
    ('ecosystem_stay',           300000),
    ('ecosystem_eat',            250000),
    ('ecosystem_experience',     250000),
    ('ecosystem_shop',           200000),
    ('ecosystem_move',           200000),
    ('ecosystem_guide',          200000)
ON CONFLICT (placement) DO NOTHING;
