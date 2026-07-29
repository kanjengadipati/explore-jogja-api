-- Add 'draft' to the partners status check constraint
ALTER TABLE partners DROP CONSTRAINT IF EXISTS chk_partners_status;
ALTER TABLE partners ADD CONSTRAINT chk_partners_status
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'suspended'));
