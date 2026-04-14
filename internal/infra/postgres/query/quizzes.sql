-- name: CreateQuiz :one
INSERT INTO quizzes (course_id, difficulty, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetQuizByID :one
SELECT * FROM quizzes WHERE id = $1;

-- name: GetQuizByCourseAndDifficulty :one
SELECT * FROM quizzes WHERE course_id = $1 AND difficulty = $2;

-- name: ListQuizzesByCourse :many
SELECT * FROM quizzes WHERE course_id = $1 ORDER BY difficulty ASC;

-- name: UpdateQuizStatus :exec
UPDATE quizzes SET status = $2, updated_at = now() WHERE id = $1;


-- name: CreateQuestion :one
INSERT INTO questions (quiz_id, type, question, choices, answer, order_idx)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListQuestionsByQuiz :many
SELECT * FROM questions WHERE quiz_id = $1 ORDER BY order_idx ASC;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: DeleteQuestionsByQuiz :exec
DELETE FROM questions WHERE quiz_id = $1;


-- name: CreateAttempt :one
INSERT INTO attempts (quiz_id, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetAttemptByID :one
SELECT * FROM attempts WHERE id = $1;

-- name: UpdateAttempt :exec
UPDATE attempts SET score = $2, total = $3, ended_at = $4
WHERE id = $1;

-- name: CreateAnswer :one
INSERT INTO answers (attempt_id, question_id, user_answer, is_correct)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAnswersByAttempt :many
SELECT * FROM answers WHERE attempt_id = $1;
