-- +goose Up
-- +goose StatementBegin
ALTER TABLE Player ADD COLUMN username VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Player DROP COLUMN IF EXISTS username;
-- +goose StatementEnd
