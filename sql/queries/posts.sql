-- name: CreatePost :execrows
INSERT INTO posts (title, link, description, published_at, feed_id)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (link)
    DO NOTHING;

-- name: GetPostsForUser :many
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
ORDER BY
    posts.published_at DESC NULLS LAST
LIMIT $2;

-- name: MarkPostRead :exec
INSERT INTO post_reads (user_id, post_id)
    VALUES ($1, $2)
ON CONFLICT (user_id, post_id)
    DO NOTHING;

-- name: MarkPostUnread :exec
DELETE FROM post_reads
WHERE user_id = $1
    AND post_id = $2;

