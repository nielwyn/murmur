-- +goose Up
CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  created_at TIMESTAMP NOT NULL DEFAULT NOW (),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW (),
  title TEXT NOT NULL,
  url TEXT UNIQUE NOT NULL,
  description TEXT,
  published_at TIMESTAMP,
  feed_id UUID NOT NULL REFERENCES feeds (id) ON DELETE CASCADE
);

CREATE INDEX idx_posts_feed_id ON posts (feed_id);

CREATE INDEX idx_posts_published_at ON posts (published_at DESC);

-- +goose Down
DROP TABLE posts;
