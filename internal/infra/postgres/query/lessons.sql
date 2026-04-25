-- name: CreateLesson :one
INSERT INTO lessons (course_id, title, order_idx)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetLessonByID :one
SELECT * FROM lessons WHERE id = $1;

-- name: ListLessonsByCourse :many
SELECT * FROM lessons WHERE course_id = $1
ORDER BY order_idx ASC;

-- name: UpdateLesson :one
UPDATE lessons SET title = $2, order_idx = $3, updated_at = now()
WHERE id = $1 RETURNING *;

-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1;
