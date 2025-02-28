-- name: CreateUser :one
INSERT INTO users (
    login,
    password_hash,
    created_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE login = $1
LIMIT 1;