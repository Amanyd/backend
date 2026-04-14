-- name: CreateCourse :one
INSERT INTO courses (title, description, rank, instructor_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;

-- name: ListCoursesByRank :many
SELECT * FROM courses WHERE rank = $1
ORDER BY created_at DESC;

-- name: ListCoursesByInstructor :many
SELECT * FROM courses WHERE instructor_id = $1
ORDER BY created_at DESC;

-- name: UpdateCourse :one
UPDATE courses
SET title = $2, description = $3, rank = $4, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;
