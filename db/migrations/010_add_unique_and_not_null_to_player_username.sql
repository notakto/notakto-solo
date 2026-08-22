-- +goose Up
-- +goose StatementBegin
ALTER TABLE Player
    ALTER COLUMN username SET NOT NULL,
    ADD CONSTRAINT player_username_unique UNIQUE (username);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Player
    DROP CONSTRAINT IF EXISTS player_username_unique,
    ALTER COLUMN username DROP NOT NULL;
-- +goose StatementEnd
