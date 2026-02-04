-- +goose Up
ALTER TABLE SessionState ADD COLUMN is_ai_move BOOLEAN[];

-- +goose Down
ALTER TABLE SessionState DROP COLUMN is_ai_move;
