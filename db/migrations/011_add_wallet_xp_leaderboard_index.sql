-- +goose Up
-- +goose StatementBegin
CREATE INDEX wallet_xp_desc_not_null_idx
    ON wallet (xp DESC)
    WHERE xp IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS wallet_xp_desc_not_null_idx;
-- +goose StatementEnd
