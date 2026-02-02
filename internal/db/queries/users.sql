-- name: CreateUser :one
INSERT INTO users (email, name, username, password_hash, key)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByKey :one
SELECT * FROM users WHERE key = $1;

-- name: UsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: UpdateUserEmailVerified :exec
UPDATE users SET email_verified = $2 WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users SET deactivated_at = NOW() WHERE id = $1;
