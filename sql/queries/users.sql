-- name: CreateUser :one
INSERT INTO users (id, username, email, created_at, updated_at)
VALUES (
    gen_random_uuid(), $1, $2, NOW(), NOW()
)
RETURNING id, username, email, created_at, updated_at;
