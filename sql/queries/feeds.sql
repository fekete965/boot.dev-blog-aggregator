-- name: CreateFeed :one
INSERT INTO feeds (id, user_id, url, name)
VALUES ($1, $2, $3, $4)
RETURNING *;
