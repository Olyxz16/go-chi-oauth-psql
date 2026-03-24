-- name: CreateUser :exec
INSERT INTO users (id, email, provider)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO NOTHING;

-- name: GetUser :one
SELECT id, email, provider
FROM users
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
