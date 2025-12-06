-- name: CreatePost :one
INSERT INTO posts (id, feed_id, title, url, description, published_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPostForUser :many
SELECT posts.*
FROM posts
INNER JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.created_at DESC
LIMIT $2;
