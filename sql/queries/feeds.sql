-- name: CreateFeed :one
INSERT INTO feeds (id, user_id, url, name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.id AS feed_id, users.name as user_name, feeds.name as feed_name, feeds.url as feed_url FROM feeds
INNER JOIN users on feeds.user_id = users.id
ORDER BY feeds.name ASC;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1 LIMIT 1;
