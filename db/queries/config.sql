-- name: GetConfigValueByKey :one
SELECT value
FROM configs
WHERE key = $1;
