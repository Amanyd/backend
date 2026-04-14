-- name: CreateFile :one
INSERT INTO files (course_id, file_name, minio_key, ingest_status)
VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files WHERE id = $1;

-- name: ListFilesByCourse :many
SELECT * FROM files WHERE course_id = $1 ORDER BY created_at DESC;

-- name: UpdateFileIngestStatus :exec
UPDATE files SET ingest_status = $2, updated_at = now()
WHERE id = $1;

-- name: AllFilesReadyForCourse :one
SELECT NOT EXISTS(
    SELECT 1 FROM files WHERE course_id = $1 AND
    ingest_status != 'ready'
) AS all_ready;
