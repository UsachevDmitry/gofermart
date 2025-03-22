-- name: CreateUser :one
INSERT INTO users (
    login,
    password
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE login = $1
LIMIT 1;