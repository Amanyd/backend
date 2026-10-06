package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/postgres/gen"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type quizRepo struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

func NewQuizRepo(pool *pgxpool.Pool) port.QuizRepository {
	return &quizRepo{pool: pool, q: gen.New(pool)}
}

// Quizzes

func (r *quizRepo) CreateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	var lessonID pgtype.UUID
	if quiz.LessonID != nil {
		lessonID = pgtype.UUID{Bytes: *quiz.LessonID, Valid: true}
	}
	row, err := r.q.CreateQuiz(ctx, gen.CreateQuizParams{
		CourseID:   quiz.CourseID,
		LessonID:   lessonID,
		Difficulty: string(quiz.Difficulty),
		Status:     string(quiz.Status),
	})
	if err != nil {
		return err
	}
	quiz.ID = row.ID
	quiz.CreatedAt = row.CreatedAt
	return nil
}

func (r *quizRepo) GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	row, err := r.q.GetQuizByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainQuiz(row), nil
}

func (r *quizRepo) GetQuizByCourseAndDifficulty(ctx context.Context, courseID uuid.UUID, difficulty domain.Difficulty) (*domain.Quiz, error) {
	row, err := r.q.GetQuizByCourseAndDifficulty(ctx, gen.GetQuizByCourseAndDifficultyParams{
		CourseID:   courseID,
		Difficulty: string(difficulty),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainQuiz(row), nil
}

func (r *quizRepo) GetQuizByLessonAndDifficulty(ctx context.Context, lessonID uuid.UUID, difficulty domain.Difficulty) (*domain.Quiz, error) {
	row, err := r.q.GetQuizByLessonAndDifficulty(ctx, gen.GetQuizByLessonAndDifficultyParams{
		LessonID:   pgtype.UUID{Bytes: lessonID, Valid: true},
		Difficulty: string(difficulty),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainQuiz(row), nil
}

func (r *quizRepo) ListQuizzesByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Quiz, error) {
	rows, err := r.q.ListQuizzesByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	quizzes := make([]domain.Quiz, len(rows))
	for i, row := range rows {
		quizzes[i] = *toDomainQuiz(row)
	}
	return quizzes, nil
}

func (r *quizRepo) ListQuizzesByCourseWithAttempts(ctx context.Context, courseID uuid.UUID, userID uuid.UUID) ([]domain.Quiz, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT q.id, q.course_id, q.lesson_id, q.difficulty, q.status, q.created_at, q.updated_at,
		       (SELECT COUNT(*) FROM questions WHERE quiz_id = q.id) AS question_count,
		       (SELECT a.score FROM attempts a WHERE a.quiz_id = q.id AND a.user_id = $2 AND a.ended_at IS NOT NULL ORDER BY a.ended_at DESC LIMIT 1) AS last_score,
		       EXISTS (SELECT 1 FROM attempts a WHERE a.quiz_id = q.id AND a.user_id = $2 AND a.ended_at IS NOT NULL) AS is_attempted
		FROM quizzes q
		WHERE q.course_id = $1
		ORDER BY q.created_at ASC
	`, courseID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quizzes := make([]domain.Quiz, 0)
	for rows.Next() {
		var q domain.Quiz
		var lessonID pgtype.UUID
		var lastScore pgtype.Float8
		if err := rows.Scan(
			&q.ID,
			&q.CourseID,
			&lessonID,
			&q.Difficulty,
			&q.Status,
			&q.CreatedAt,
			&q.UpdatedAt,
			&q.QuestionCount,
			&lastScore,
			&q.IsAttempted,
		); err != nil {
			return nil, err
		}
		if lessonID.Valid {
			id := uuid.UUID(lessonID.Bytes)
			q.LessonID = &id
		}
		if lastScore.Valid {
			score := lastScore.Float64
			q.LastScore = &score
		}
		quizzes = append(quizzes, q)
	}
	return quizzes, nil
}

func (r *quizRepo) UpdateQuizStatus(ctx context.Context, id uuid.UUID, status domain.QuizStatus) error {
	return r.q.UpdateQuizStatus(ctx, gen.UpdateQuizStatusParams{
		ID:     id,
		Status: string(status),
	})
}

func (r *quizRepo) DeleteQuizzesByCourse(ctx context.Context, courseID uuid.UUID) error {
	return r.q.DeleteQuizzesByCourse(ctx, courseID)
}

func toDomainQuiz(q gen.Quiz) *domain.Quiz {
	quiz := &domain.Quiz{
		ID:         q.ID,
		CourseID:   q.CourseID,
		Difficulty: domain.Difficulty(q.Difficulty),
		Status:     domain.QuizStatus(q.Status),
		CreatedAt:  q.CreatedAt,
		UpdatedAt:  q.UpdatedAt,
	}
	if q.LessonID.Valid {
		id := uuid.UUID(q.LessonID.Bytes)
		quiz.LessonID = &id
	}
	return quiz
}

// Questions

func (r *quizRepo) CreateQuestions(ctx context.Context, questions []domain.Question) error {
	for _, q := range questions {
		choicesJSON, err := json.Marshal(q.Choices)
		if err != nil {
			return err
		}
		_, err = r.pool.Exec(ctx, `
			INSERT INTO questions (quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, q.QuizID, string(q.Type), q.Question, choicesJSON, q.Answer, int32(q.OrderIdx), q.Explanation, q.Difficulty, q.TopicPhrase)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *quizRepo) ListQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase
		FROM questions
		WHERE quiz_id = $1
		ORDER BY order_idx ASC
	`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

func (r *quizRepo) SampleQuestionsByQuiz(ctx context.Context, quizID uuid.UUID, limit int) ([]domain.Question, error) {
	// Pick 1 easy question per topic from the question bank.
	// If limit > 0, cap the total questions returned; otherwise return 1 per topic for all topics.
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}
	query := fmt.Sprintf(`
		WITH ranked AS (
			SELECT id, quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase,
			       ROW_NUMBER() OVER (
			           PARTITION BY COALESCE(NULLIF(topic_phrase, ''), id::text)
			           ORDER BY (CASE WHEN difficulty = 'easy' THEN 0 WHEN difficulty = 'medium' THEN 1 ELSE 2 END), RANDOM()
			       ) as rn
			FROM questions
			WHERE quiz_id = $1
		)
		SELECT id, quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase
		FROM ranked
		WHERE rn = 1
		ORDER BY RANDOM()
		%s
	`, limitClause)

	rows, err := r.pool.Query(ctx, query, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanQuestions(rows)
}

func (r *quizRepo) SampleQuestionsByCourse(ctx context.Context, courseID uuid.UUID, difficulty domain.Difficulty, limit int) ([]domain.Question, error) {
	// Filter questions by difficulty across all lessons in the course:
	// "medium" for Basic course assessment, "hard" for Advanced course assessment.
	limitClause := ""
	if limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", limit)
	}
	query := fmt.Sprintf(`
		WITH filtered AS (
			SELECT q.id, q.quiz_id, q.type, q.question, q.choices, q.answer, q.order_idx, q.explanation, q.difficulty, q.topic_phrase
			FROM questions q
			JOIN quizzes qz ON qz.id = q.quiz_id
			WHERE qz.course_id = $1 AND ($2 = '' OR q.difficulty = $2)
		)
		SELECT id, quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase
		FROM filtered
		ORDER BY RANDOM()
		%s
	`, limitClause)

	rows, err := r.pool.Query(ctx, query, courseID, string(difficulty))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions, err := r.scanQuestions(rows)
	if err != nil {
		return nil, err
	}

	// Fallback if no questions matched the specific difficulty
	if len(questions) == 0 && difficulty != "" {
		fallbackQuery := fmt.Sprintf(`
			SELECT q.id, q.quiz_id, q.type, q.question, q.choices, q.answer, q.order_idx, q.explanation, q.difficulty, q.topic_phrase
			FROM questions q
			JOIN quizzes qz ON qz.id = q.quiz_id
			WHERE qz.course_id = $1
			ORDER BY RANDOM()
			%s
		`, limitClause)
		fallbackRows, err := r.pool.Query(ctx, fallbackQuery, courseID)
		if err != nil {
			return nil, err
		}
		defer fallbackRows.Close()
		return r.scanQuestions(fallbackRows)
	}

	return questions, nil
}

func (r *quizRepo) scanQuestions(rows pgx.Rows) ([]domain.Question, error) {
	questions := make([]domain.Question, 0)
	for rows.Next() {
		var q domain.Question
		var choicesBytes []byte
		var explanation, difficulty, topicPhrase pgtype.Text
		if err := rows.Scan(
			&q.ID,
			&q.QuizID,
			&q.Type,
			&q.Question,
			&choicesBytes,
			&q.Answer,
			&q.OrderIdx,
			&explanation,
			&difficulty,
			&topicPhrase,
		); err != nil {
			return nil, err
		}
		if len(choicesBytes) > 0 {
			_ = json.Unmarshal(choicesBytes, &q.Choices)
		}
		if explanation.Valid {
			q.Explanation = explanation.String
		}
		if difficulty.Valid {
			q.Difficulty = difficulty.String
		}
		if topicPhrase.Valid {
			q.TopicPhrase = topicPhrase.String
		}
		questions = append(questions, q)
	}
	return questions, nil
}

func (r *quizRepo) GetQuestionByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	var q domain.Question
	var choicesBytes []byte
	var explanation, difficulty, topicPhrase pgtype.Text
	err := r.pool.QueryRow(ctx, `
		SELECT id, quiz_id, type, question, choices, answer, order_idx, explanation, difficulty, topic_phrase
		FROM questions
		WHERE id = $1
	`, id).Scan(
		&q.ID,
		&q.QuizID,
		&q.Type,
		&q.Question,
		&choicesBytes,
		&q.Answer,
		&q.OrderIdx,
		&explanation,
		&difficulty,
		&topicPhrase,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	if len(choicesBytes) > 0 {
		_ = json.Unmarshal(choicesBytes, &q.Choices)
	}
	if explanation.Valid {
		q.Explanation = explanation.String
	}
	if difficulty.Valid {
		q.Difficulty = difficulty.String
	}
	if topicPhrase.Valid {
		q.TopicPhrase = topicPhrase.String
	}
	return &q, nil
}

func (r *quizRepo) DeleteQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) error {
	return r.q.DeleteQuestionsByQuiz(ctx, quizID)
}

// Attempts

func (r *quizRepo) CreateAttempt(ctx context.Context, attempt *domain.Attempt) error {
	row, err := r.q.CreateAttempt(ctx, gen.CreateAttemptParams{
		QuizID: attempt.QuizID,
		UserID: attempt.UserID,
	})
	if err != nil {
		return err
	}
	attempt.ID = row.ID
	attempt.StartedAt = row.StartedAt
	return nil
}

func (r *quizRepo) GetAttemptByID(ctx context.Context, id uuid.UUID) (*domain.Attempt, error) {
	row, err := r.q.GetAttemptByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainAttempt(row), nil
}

func (r *quizRepo) UpdateAttempt(ctx context.Context, attempt *domain.Attempt) error {
	var endedAt pgtype.Timestamptz
	if attempt.EndedAt != nil {
		endedAt = pgtype.Timestamptz{Time: *attempt.EndedAt, Valid: true}
	}
	return r.q.UpdateAttempt(ctx, gen.UpdateAttemptParams{
		ID:      attempt.ID,
		Score:   attempt.Score,
		Total:   int32(attempt.Total),
		EndedAt: endedAt,
	})
}

func toDomainAttempt(a gen.Attempt) *domain.Attempt {
	attempt := &domain.Attempt{
		ID:        a.ID,
		QuizID:    a.QuizID,
		UserID:    a.UserID,
		Score:     a.Score,
		Total:     int(a.Total),
		StartedAt: a.StartedAt,
	}
	if a.EndedAt.Valid {
		attempt.EndedAt = &a.EndedAt.Time
	}
	return attempt
}

// Answers

func (r *quizRepo) CreateAnswer(ctx context.Context, answer *domain.Answer) error {
	row, err := r.q.CreateAnswer(ctx, gen.CreateAnswerParams{
		AttemptID:  answer.AttemptID,
		QuestionID: answer.QuestionID,
		UserAnswer: answer.UserAnswer,
		IsCorrect:  answer.IsCorrect,
	})
	if err != nil {
		return err
	}
	answer.ID = row.ID
	return nil
}

func (r *quizRepo) ListAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]domain.Answer, error) {
	rows, err := r.q.ListAnswersByAttempt(ctx, attemptID)
	if err != nil {
		return nil, err
	}
	answers := make([]domain.Answer, len(rows))
	for i, row := range rows {
		answers[i] = domain.Answer{
			ID:         row.ID,
			AttemptID:  row.AttemptID,
			QuestionID: row.QuestionID,
			UserAnswer: row.UserAnswer,
			IsCorrect:  row.IsCorrect,
		}
	}
	return answers, nil
}
