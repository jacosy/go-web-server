-- name: CreateUser :one
INSERT INTO users (id, username, email, hashed_password, created_at, updated_at)
VALUES (
    gen_random_uuid(), $1, $2, $3, NOW(), NOW()
)
RETURNING id, username, email, created_at, updated_at;

-- name: Reset :exec
DELETE FROM users;

-- name: TruncateUsers :exec
TRUNCATE TABLE users RESTART IDENTITY;

