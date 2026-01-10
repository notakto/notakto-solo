-- name: CreateWallet :exec
INSERT INTO wallet ( uid, coins, xp)
VALUES ($1, $2, $3);

-- name: GetWalletByPlayerId :one
SELECT
    uid,
    coins,
    xp
FROM wallet
WHERE uid = $1;

-- name: UpdateWalletCoinsAndXpReward :exec
UPDATE wallet
SET coins = coins+$2,
    xp = xp+$3
WHERE uid = $1;

-- name: UpdateWalletXpReward :exec
UPDATE wallet
SET xp = xp+$2
WHERE uid = $1;

-- name: UpdateWalletReduceCoins :exec
UPDATE wallet
SET coins = coins-$2
WHERE uid = $1;

-- name: GetWalletByPlayerIdWithLock :one
SELECT
    uid,
    coins,
    xp
FROM wallet
WHERE uid = $1
FOR UPDATE;