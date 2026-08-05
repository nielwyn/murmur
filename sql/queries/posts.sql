-- name: CreatePost :execrows
INSERT INTO posts (title, link, description, published_at, feed_id, image_url)
    VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (link)
    DO NOTHING;

-- name: GetPostsForUser :many
-- Keyset pagination via before_id/before_published_at (the last post's own
-- fields). NULLs coalesce to -infinity since NULLS LAST sorts them oldest.
-- unread/feed_title filter here, not client-side, so pagination stays correct.
SELECT
    posts.*,
    feeds.title AS feed_title,
    (post_reads.read_at IS NOT NULL)::boolean AS read
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON feeds.id = posts.feed_id
    LEFT JOIN post_reads ON post_reads.post_id = posts.id
        AND post_reads.user_id = $1
WHERE
    feed_follows.user_id = $1
    -- pagination cursor
    AND (
        sqlc.narg('before_id')::uuid IS NULL
        OR (COALESCE(posts.published_at, '-infinity'), posts.id)
            < (COALESCE(sqlc.narg('before_published_at')::timestamp, '-infinity'), sqlc.narg('before_id')::uuid)
    )
    -- unread filter
    AND (sqlc.narg('unread')::boolean IS NOT TRUE OR post_reads.read_at IS NULL)
    -- feed filter
    AND (sqlc.narg('feed_title')::text IS NULL OR feeds.title = sqlc.narg('feed_title')::text)
ORDER BY
    posts.published_at DESC NULLS LAST, posts.id DESC
LIMIT $2;

-- name: CountUnreadPostsForUser :one
SELECT
    count(*)
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    LEFT JOIN post_reads ON post_reads.post_id = posts.id
        AND post_reads.user_id = $1
WHERE
    feed_follows.user_id = $1
    AND post_reads.read_at IS NULL;

-- name: MarkPostRead :exec
INSERT INTO post_reads (user_id, post_id)
    VALUES ($1, $2)
ON CONFLICT (user_id, post_id)
    DO NOTHING;

-- name: MarkPostUnread :exec
DELETE FROM post_reads
WHERE user_id = $1
    AND post_id = $2;

