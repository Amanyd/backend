-- name: CreateCourse :one
INSERT INTO courses (title, description, rank, instructor_id, published)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;

-- name: ListCoursesByRank :many
SELECT * FROM courses WHERE rank = $1 AND published = true
ORDER BY created_at DESC;

-- name: ListCoursesByInstructor :many
SELECT * FROM courses WHERE instructor_id = $1
ORDER BY created_at DESC;

-- name: UpdateCourse :one
UPDATE courses
SET title = $2, description = $3, rank = $4, published = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: PublishCourse :exec
UPDATE courses SET published = true, updated_at = now() WHERE id = $1;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
