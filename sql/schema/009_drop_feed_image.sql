-- +goose Up
ALTER TABLE feeds
    DROP COLUMN image_url;

-- +goose Down
ALTER TABLE feeds
    ADD COLUMN image_url text;
