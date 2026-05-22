ALTER TABLE refresh_tokens
    ADD COLUMN family_id VARCHAR(255),
    ADD COLUMN rotated_from_token_id BIGINT UNSIGNED,
    ADD COLUMN replaced_by_token_id BIGINT UNSIGNED,
    ADD COLUMN revoked_at DATETIME,
    ADD COLUMN revoke_reason TEXT,
    ADD CONSTRAINT fk_refresh_tokens_rotated_from FOREIGN KEY (rotated_from_token_id) REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    ADD CONSTRAINT fk_refresh_tokens_replaced_by FOREIGN KEY (replaced_by_token_id) REFERENCES refresh_tokens(id) ON DELETE SET NULL;

UPDATE refresh_tokens
SET family_id = token_hash
WHERE family_id IS NULL OR family_id = '';

ALTER TABLE refresh_tokens
    MODIFY COLUMN family_id VARCHAR(255) NOT NULL;

CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
CREATE INDEX idx_refresh_tokens_revoked_at ON refresh_tokens(revoked_at);
CREATE INDEX idx_refresh_tokens_active_user_id ON refresh_tokens(user_id, revoked_at, deleted_at);
