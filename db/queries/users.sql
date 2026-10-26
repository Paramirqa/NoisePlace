-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password_hash
)
VALUES (
    sqlc.arg(username),
    sqlc.arg(email),
    sqlc.arg(password_hash)
)
RETURNING
    id,
    username,
    email,
    created_at,
    updated_at;


-- name: GetUserByID :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE id = sqlc.arg(id)
LIMIT 1;


-- name: GetUserByEmail :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE email = sqlc.arg(email)
LIMIT 1;


-- name: GetUserByUsername :one
SELECT
    id,
    username,
    email,
    password_hash,
    created_at,
    updated_at
FROM users
WHERE username = sqlc.arg(username)
LIMIT 1;


-- name: UpdateUser :one
UPDATE users
SET
    username = sqlc.arg(username),
    email = sqlc.arg(email),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING
    id,
    username,
    email,
    created_at,
    updated_at;