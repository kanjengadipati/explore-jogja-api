DROP INDEX IF EXISTS idx_sales_commissions_tier;
ALTER TABLE sales_commissions DROP COLUMN IF EXISTS tier;
