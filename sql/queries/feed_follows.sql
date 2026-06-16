-- name: CreateFeedFollow :one
INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
RETURNING
    *;

-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.*,
    feeds.name AS feed_name,
    feeds.url AS feed_url
FROM
    feed_follows
    JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE
    feed_follows.user_id = $1
ORDER BY
    feed_follows.created_at;

-- name: DeleteFeedFollow :execrows
DELETE FROM feed_follows
WHERE user_id = $1
    AND feed_id = $2;

