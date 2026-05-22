DROP INDEX idx_refresh_tokens_active_user_id ON refresh_tokens;
DROP INDEX idx_refresh_tokens_revoked_at ON refresh_tokens;
DROP INDEX idx_refresh_tokens_family_id ON refresh_tokens;

ALTER TABLE refresh_tokens
    DROP FOREIGN KEY fk_refresh_tokens_rotated_from,
    DROP FOREIGN KEY fk_refresh_tokens_replaced_by;

ALTER TABLE refresh_tokens
    DROP COLUMN revoke_reason,
    DROP COLUMN revoked_at,
    DROP COLUMN replaced_by_token_id,
    DROP COLUMN rotated_from_token_id,
    DROP COLUMN family_id;
