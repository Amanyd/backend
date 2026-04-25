-- name: CreateFile :one
INSERT INTO files (lesson_id, file_name, file_type, minio_key, ingest_status)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files WHERE id = $1;

-- name: ListFilesByLesson :many
SELECT * FROM files WHERE lesson_id = $1 ORDER BY created_at DESC;

-- name: UpdateFileIngestStatus :exec
UPDATE files SET ingest_status = $2, updated_at = now() WHERE id = $1;

-- name: AllFilesReadyForCourse :one
SELECT NOT EXISTS (
    SELECT 1 FROM files f
    JOIN lessons l ON l.id = f.lesson_id
    WHERE l.course_id = $1 AND f.ingest_status != 'ready'
) AS all_ready;
