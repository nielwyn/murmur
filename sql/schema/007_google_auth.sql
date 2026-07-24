-- +goose Up
ALTER TABLE users
    ALTER COLUMN hashed_password DROP NOT NULL;

ALTER TABLE users
    ADD COLUMN google_id text UNIQUE;

-- +goose Down
ALTER TABLE users
    DROP COLUMN google_id;

ALTER TABLE users
    ALTER COLUMN hashed_password SET NOT NULL;
