-- +goose Up
-- Add the column with a default of empty array
ALTER TABLE SessionState ADD COLUMN is_ai_move BOOLEAN[] DEFAULT '{}';

-- Backfill existing rows: set is_ai_move to array of false with same length as boards
UPDATE SessionState
SET is_ai_move = CASE
    WHEN boards IS NULL OR array_length(boards, 1) IS NULL THEN '{}'::BOOLEAN[]
    ELSE array_fill(FALSE, ARRAY[array_length(boards, 1)])
END;

-- +goose Down
ALTER TABLE SessionState DROP COLUMN is_ai_move;
