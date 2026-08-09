-- Rename tenant_user_id -> partner_user_id (naming consistency with sales_commissions).
ALTER INDEX IF EXISTS idx_sales_bonuses_onboarding_tenant RENAME TO idx_sales_bonuses_onboarding_partner;
ALTER TABLE sales_bonuses RENAME COLUMN tenant_user_id TO partner_user_id;
