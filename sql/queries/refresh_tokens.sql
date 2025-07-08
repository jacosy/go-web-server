-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, user_id, created_at, updated_at, expires_at)
VALUES (
    $1, $2, NOW(), NOW(), NOW() + INTERVAL '60 days'
);
