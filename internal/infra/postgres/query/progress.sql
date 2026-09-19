-- name: GetCourseProgress :many
SELECT * FROM user_course_progress
WHERE user_id = $1 AND course_id = $2;

-- name: GetLessonProgress :many
SELECT ulp.* FROM user_lesson_progress ulp
JOIN lessons l ON l.id = ulp.lesson_id
WHERE ulp.user_id = $1 AND l.course_id = $2;

-- name: GetFileProgress :many
SELECT ufp.*, f.lesson_id FROM user_file_progress ufp
JOIN files f ON f.id = ufp.file_id
JOIN lessons l ON l.id = f.lesson_id
WHERE ufp.user_id = $1 AND l.course_id = $2;

-- name: UpsertFileProgress :one
INSERT INTO user_file_progress (user_id, file_id, is_viewed)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, file_id) DO UPDATE SET is_viewed = $3
RETURNING *;

-- name: UpsertLessonProgress :one
INSERT INTO user_lesson_progress (user_id, lesson_id, is_completed, completed_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, lesson_id) DO UPDATE SET 
is_completed = $3, completed_at = $4
RETURNING *;

-- name: UpsertCourseProgress :one
INSERT INTO user_course_progress (user_id, course_id, is_completed, completed_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, course_id) DO UPDATE SET 
is_completed = $3, completed_at = $4
RETURNING *;
