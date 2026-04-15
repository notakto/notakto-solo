-- name: CreatePayment :exec
INSERT INTO Payment (id, uid, package_id, coins, amount_cents, status, hosted_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW());

-- name: GetPaymentById :one
SELECT id, uid, package_id, coins, amount_cents, status, hosted_url, created_at, updated_at
FROM Payment
WHERE id = $1;

-- name: GetPaymentByIdWithLock :one
SELECT id, uid, package_id, coins, amount_cents, status, hosted_url, created_at, updated_at
FROM Payment
WHERE id = $1
FOR UPDATE;

-- name: UpdatePaymentStatusIfNotConfirmed :execrows
UPDATE Payment
SET status = $2, updated_at = NOW()
WHERE id = $1 AND status != 'confirmed';

-- name: GetPaymentsByUid :many
SELECT id, uid, package_id, coins, amount_cents, status, hosted_url, created_at, updated_at
FROM Payment
WHERE uid = $1
ORDER BY created_at DESC;
