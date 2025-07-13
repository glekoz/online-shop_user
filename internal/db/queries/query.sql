-- name: Get :one
SELECT *
FROM users
WHERE id = $1;

-- name: Save :one
INSERT INTO users(name, email, hashed_password, created_at)
VALUES ($1, $2, $3, NOW())
RETURNING id;