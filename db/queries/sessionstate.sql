-- name: CreateInitialSessionState :exec
INSERT INTO sessionstate (session_id, boards, is_ai_move)
VALUES ($1, $2, $3);

-- name: UpdateSessionState :exec
UPDATE sessionstate
SET boards = $2, is_ai_move = $3
WHERE session_id = $1;