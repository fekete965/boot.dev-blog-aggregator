-- name: CreateUser :one
INSERT INTO users (id, name, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: FindUserByeName :one
SELECT * FROM users WHERE name = $1 ORDER BY created_at DESC LIMIT 1;

-- name: DeleteAllUsers :exec
DELETE FROM users;
