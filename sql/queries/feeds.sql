-- name: CreateFeed :one
INSERT INTO feeds (title, link, user_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetFeedByLink :one
SELECT
    *
FROM
    feeds
WHERE
    link = $1;

-- name: GetFeeds :many
SELECT
    feeds.*,
    users.username AS creator_name
FROM
    feeds
    JOIN users ON feeds.user_id = users.id
ORDER BY
    feeds.created_at;

-- name: GetFeedsDueForFetch :many
SELECT
    *
FROM
    feeds
WHERE
    last_fetched_at IS NULL
    OR last_fetched_at < NOW() - (fetch_interval_seconds * INTERVAL '1 second')
ORDER BY
    last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedFetched :one
UPDATE
    feeds
SET
    last_fetched_at = NOW(),
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    *;

