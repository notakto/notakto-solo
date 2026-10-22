-- +goose Up
-- +goose StatementBegin
ALTER TABLE Wallet
    ALTER COLUMN coins SET NOT NULL,
    ALTER COLUMN xp SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE Wallet
    ALTER COLUMN coins DROP NOT NULL,
    ALTER COLUMN xp DROP NOT NULL;
-- +goose StatementEnd
