-- name: GetPlayerById :one
SELECT * FROM Player WHERE uid = $1;

-- name: GetPlayerByIdWithLock :one
SELECT * FROM Player WHERE uid = $1 FOR UPDATE;

-- name: CreatePlayer :exec
INSERT INTO Player (uid, email, name, profile_pic, username)
VALUES ($1, $2, $3, $4, $5);

-- name: UpdatePlayerName :one
UPDATE Player SET name = $2 WHERE uid = $1 RETURNING *;

-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM Player WHERE username = $1 AND uid != $2) AS exists;

-- name: UpdatePlayerUsername :one
UPDATE Player SET username = $2 WHERE uid = $1 RETURNING *;

-- name: UpdatePlayerProfileImage :one
UPDATE Player
SET profile_image_file_id = $2,
    profile_image_file_path = $3
WHERE uid = $1
RETURNING *;
