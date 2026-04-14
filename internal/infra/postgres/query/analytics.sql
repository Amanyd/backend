-- name: RecordEvent :one
INSERT INTO events (user_id, course_id, type, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;


-- name: GetCourseMetrics :one
SELECT
    $1::uuid AS course_id,
    (SELECT COUNT(DISTINCT a.user_id) FROM attempts a
     JOIN quizzes q ON q.id = a.quiz_id
     WHERE q.course_id = $1) AS total_students,
    (SELECT COALESCE(AVG(a.score), 0) FROM attempts a
     JOIN quizzes q ON q.id = a.quiz_id
     WHERE q.course_id = $1 AND a.ended_at IS NOT NULL) AS avg_quiz_score,
    (SELECT COUNT(*) FROM chat_messages cm
     JOIN chat_sessions cs ON cs.id = cm.session_id
     WHERE cs.course_id = $1) AS total_messages,
    (SELECT COUNT(*) FROM files
     WHERE course_id = $1) AS total_files;

-- name: GetStudentScores :many
SELECT
    u.id AS user_id,
    u.name,
    u.rank,
    COALESCE(AVG(a.score), 0)::float AS avg_score
FROM users u
JOIN attempts a ON a.user_id = u.id
JOIN quizzes q ON q.id = a.quiz_id
WHERE q.course_id = $1 AND a.ended_at IS NOT NULL
GROUP BY u.id, u.name, u.rank
ORDER BY avg_score DESC;

-- name: GetOverview :one
SELECT
    (SELECT COUNT(DISTINCT a.user_id) FROM attempts a
     JOIN quizzes q ON q.id = a.quiz_id
     JOIN courses c ON c.id = q.course_id
     WHERE c.instructor_id = $1) AS total_students,
    (SELECT COUNT(*) FROM courses
     WHERE instructor_id = $1) AS total_courses,
    (SELECT COALESCE(AVG(a.score), 0) FROM attempts a
     JOIN quizzes q ON q.id = a.quiz_id
     JOIN courses c ON c.id = q.course_id
     WHERE c.instructor_id = $1 AND a.ended_at IS NOT NULL) AS avg_score;
