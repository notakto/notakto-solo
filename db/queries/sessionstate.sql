-- name: CreateInitialSessionState :exec
INSERT INTO sessionstate (session_id, boards)
VALUES ($1, $2);

-- name: UpdateSessionState :exec
UPDATE sessionstate
SET boards = $2
WHERE session_id = $1;