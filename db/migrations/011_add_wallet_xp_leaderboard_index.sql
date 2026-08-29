-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY wallet_xp_desc_not_null_idx
    ON wallet (xp DESC)
    WHERE xp IS NOT NULL;

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS wallet_xp_desc_not_null_idx;
