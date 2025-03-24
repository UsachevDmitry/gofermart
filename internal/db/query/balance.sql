-- name: GetBalanceByUserID :one
SELECT current_balance, withdrawn_balance
FROM loyalty_accounts
WHERE user_id = $1
LIMIT 1;

-- name: CreateWithdrawal :exec
INSERT INTO withdrawals (user_id, order_number, sum)
VALUES ($1, $2, $3);

-- name: UpdateBalance :exec
UPDATE loyalty_accounts
SET current_balance = $2, withdrawn_balance = $3
WHERE user_id = $1;

-- name: GetWithdrawalsByUserID :many
SELECT 
    order_number, 
    user_id, 
    sum, 
    processed_at
FROM withdrawals
WHERE 
    user_id = $1
ORDER BY processed_at DESC;