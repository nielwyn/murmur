-- name: CreateUser :one
INSERT INTO users (name, email, hashed_password)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
