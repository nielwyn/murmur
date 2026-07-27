-- +goose Up
ALTER TABLE feeds
    ADD COLUMN image_url text;

ALTER TABLE posts
    ADD COLUMN image_url text;

-- +goose Down
ALTER TABLE feeds
    DROP COLUMN image_url;

ALTER TABLE posts
    DROP COLUMN image_url;
