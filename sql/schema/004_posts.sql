-- +goose Up
CREATE TABLE posts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW(),
    title text NOT NULL,
    url text UNIQUE NOT NULL,
    description text,
    published_at timestamp,
    feed_id uuid NOT NULL REFERENCES feeds (id) ON DELETE CASCADE
);

CREATE INDEX idx_posts_feed_id ON posts (feed_id);

CREATE INDEX idx_posts_published_at ON posts (published_at DESC);

-- +goose Down
DROP TABLE posts;

