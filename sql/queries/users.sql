-- name: CreateUser :one
INSERT INTO users (username, email, hashed_password)
    VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetUserByName :one
SELECT
    *
FROM
    users
WHERE
    username = $1;

-- name: GetUserByID :one
SELECT
    *
FROM
    users
WHERE
    id = $1;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;

-- name: GetUserByGoogleID :one
SELECT
    *
FROM
    users
WHERE
    google_id = $1;

-- name: LinkGoogleAccount :one
UPDATE
    users
SET
    google_id = $1,
    updated_at = NOW()
WHERE
    id = $2
RETURNING
    *;

-- name: CreateGoogleUser :one
INSERT INTO users (username, email, google_id)
    VALUES ($1, $2, $3)
RETURNING
    *;

