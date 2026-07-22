-- name: CreatePost :execrows
INSERT INTO posts (title, link, description, published_at, feed_id)
    VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (link)
    DO NOTHING;

-- name: GetPostsForUser :many
SELECT
    posts.*,
    feeds.title AS feed_title
FROM
    posts
    JOIN feed_follows ON feed_follows.feed_id = posts.feed_id
    JOIN feeds ON feeds.id = posts.feed_id
WHERE
    feed_follows.user_id = $1
ORDER BY
    posts.published_at DESC NULLS LAST
LIMIT $2;

