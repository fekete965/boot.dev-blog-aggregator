WITH inserted_feed_follows AS (
  INSERT INTO feed_follows (id,
  user_id, feed_id, created_at)
  VALUES ($1, $2, $3, $4)
  RETURNING *
)

-- name: CreateFeedFollow :one
SELECT 
  inserted_feed_follows.*,
  feeds.name as feed_name,
  users.name as user_name
FROM inserted_feed_follows
INNER JOIN feeds on inserted_feed_follows.feed_id = feeds.id
INNER JOIN users on inserted_feed_follows.user_id = users.id
ORDER BY inserted_feed_follows.created_at DESC
LIMIT 1;

-- name: GetFeedFollowsForUser :many
SELECT 
  feed_follows.id AS feed_follow_id,
  feeds.name as feed_name,
  feeds.url as feed_url
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1 
ORDER BY feed_follows.created_at ASC;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE
  user_id = $1 AND feed_id = $2;
