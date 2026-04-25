-- name: CreateUser :one
INSERT INTO users (name, enrollment_id, rank, role, password_hash)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEnrollmentID :one
SELECT * FROM users WHERE enrollment_id = $1;
