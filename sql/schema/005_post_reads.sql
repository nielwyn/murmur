-- +goose Up
CREATE TABLE post_reads (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    post_id uuid NOT NULL REFERENCES posts (id) ON DELETE CASCADE,
    read_at timestamp NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

-- +goose Down
DROP TABLE post_reads;

