-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    created_at timestamp NOT NULL DEFAULT NOW(),
    updated_at timestamp NOT NULL DEFAULT NOW(),
    name text UNIQUE NOT NULL,
    email text UNIQUE NOT NULL,
    hashed_password text NOT NULL
);

-- +goose Down
DROP TABLE users;

