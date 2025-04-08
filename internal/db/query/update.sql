-- name: UpdateOrderStatus :exec
UPDATE orders
SET
    status = $1,
    accrual = $2
WHERE
    order_number = $3;

-- name: GetOrderOwner :one
SELECT user_id
FROM orders
WHERE order_number = $1;

-- name: UpdateLoyaltyAccountBalance :exec
UPDATE loyalty_accounts
SET
    current_balance = current_balance + $1,
    updated_at = NOW()
WHERE
    user_id = $2;

-- name: CreateLoyaltyTransaction :exec
INSERT INTO loyalty_transactions (
    account_id,
    order_id,
    points,
    transaction_type,
    created_at
)
VALUES (
    $1, $2, $3, $4, NOW()
);

-- name: GetLoyaltyAccountID :one
SELECT id
FROM loyalty_accounts
WHERE user_id = $1;

-- name: GetOrderID :one
SELECT id
FROM orders
WHERE order_number = $1;

-- name: SaveOrder :exec
INSERT INTO orders (
    user_id,
    order_number,
    status,
    accrual,
    uploaded_at
)
VALUES (
    $1, $2, $3, $4, NOW()
);

-- -- name: GetLoyaltyAccountForUpdate :one
-- SELECT * FROM loyalty_accounts 
-- WHERE user_id = $1 
-- FOR UPDATE;
