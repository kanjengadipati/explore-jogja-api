-- Revert: remove 'draft' from partners status constraint
ALTER TABLE partners DROP CONSTRAINT IF EXISTS chk_partners_status;
ALTER TABLE partners ADD CONSTRAINT chk_partners_status
    CHECK (status IN ('pending', 'approved', 'rejected', 'suspended'));
