-- name: GetPlayerById :one
SELECT * FROM Player WHERE uid = $1;

-- name: CreatePlayer :exec
INSERT INTO Player (uid, email, name, profile_pic, username)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePlayerName :one
UPDATE Player SET name = $2 WHERE uid = $1 RETURNING *;