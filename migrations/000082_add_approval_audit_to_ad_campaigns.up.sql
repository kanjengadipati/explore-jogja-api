-- migration: 000082_add_approval_audit_to_ad_campaigns
-- Audit trail for the mandatory ad-campaign approval flow. The approval state
-- itself lives in payment_status (pending_review → pending_payment/rejected);
-- these columns record WHO acted and WHEN, plus the rejection reason that is
-- surfaced to the business owner in their dashboard and in the rejection email.

ALTER TABLE ad_campaigns
    ADD COLUMN IF NOT EXISTS approved_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS approved_by     VARCHAR(255),
    ADD COLUMN IF NOT EXISTS rejection_reason TEXT,
    ADD COLUMN IF NOT EXISTS rejected_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS rejected_by     VARCHAR(255);
