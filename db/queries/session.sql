-- name: GetLatestSessionStateByPlayerId :one
SELECT 
    s.session_id,
    s.uid,
    s.created_at,
    s.gameover,
    s.winner,
    s.board_size,
    s.number_of_boards,
    s.difficulty,
    ss.boards
FROM session s
JOIN sessionstate ss
    ON s.session_id = ss.session_id
WHERE s.uid = $1
    AND s.created_at >= now() - interval '15 minutes'
ORDER BY s.created_at DESC
LIMIT 1;

-- name: CreateSession :exec
INSERT INTO session (session_id, uid, created_at, gameover, winner, board_size, number_of_boards, difficulty)
VALUES ($1, $2, now(), false, NULL, $3, $4, $5);

-- name: UpdateSessionAfterGameover :exec
UPDATE session
SET gameover = true,
    winner = $2
WHERE session_id = $1;