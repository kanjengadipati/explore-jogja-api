ALTER TABLE sales_commissions ADD COLUMN IF NOT EXISTS tier INTEGER NOT NULL DEFAULT 1;
CREATE INDEX IF NOT EXISTS idx_sales_commissions_tier ON sales_commissions(tier);
