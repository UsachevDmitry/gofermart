-- name: CreateOrder :one
INSERT INTO orders (
    user_id, 
    order_number, 
    status
    ) VALUES ($1, $2, $3)
RETURNING *;

-- name: GetOrderByNumber :one
SELECT * FROM orders
WHERE order_number = $1 LIMIT 1;

-- name: GetOrdersByUserID :many
SELECT order_number, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1
ORDER BY uploaded_at DESC;

-- name: GetUnprocessedOrders :many
SELECT id, order_number, status 
FROM orders 
WHERE status NOT IN ('PROCESSED', 'INVALID');

-- name: GetProcessedOrdersByUserID :many
SELECT order_number, user_id, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1 AND status = 'PROCESSED'
ORDER BY uploaded_at DESC;