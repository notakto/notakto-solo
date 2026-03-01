-- +goose Up
-- +goose StatementBegin
CREATE TABLE Payment (
    id TEXT PRIMARY KEY,
    uid VARCHAR(36) NOT NULL,
    package_id TEXT NOT NULL,
    coins INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'created',
    hosted_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (uid) REFERENCES Player(uid) ON DELETE CASCADE
);

CREATE INDEX idx_payment_uid ON Payment(uid);
CREATE INDEX idx_payment_status ON Payment(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payment_status;
DROP INDEX IF EXISTS idx_payment_uid;
DROP TABLE IF EXISTS Payment;
-- +goose StatementEnd
