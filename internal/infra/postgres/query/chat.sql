-- name: CreateChatSession :one
INSERT INTO chat_sessions (user_id,course_id,title)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetChatSessionByID :one
SELECT * from chat_sessions WHERE id = $1;

-- name: ListChatSessionsByUser :many
SELECT * FROM chat_sessions WHERE user_id = $1 ORDER BY updated_at DESC;

-- name: CreateMessage :one
INSERT INTO chat_messages (session_id, role, content, citations)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListMessagesBySession :many
SELECT * FROM (
    SELECT * FROM chat_messages
    WHERE session_id = $1
    ORDER BY created_at DESC
    LIMIT $2
) sub ORDER BY created_at ASC;
