package postgres

import (
	"context"
	"encoding/json"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/eulerbutcooler/wingman-backend/internal/infra/postgres/gen"
	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type quizRepo struct {
	q *gen.Queries
}

func NewQuizRepo(pool *pgxpool.Pool) port.QuizRepository {
	return &quizRepo{q: gen.New(pool)}
}

// Quizzes

func (r *quizRepo) CreateQuiz(ctx context.Context, quiz *domain.Quiz) error {
	row, err := r.q.CreateQuiz(ctx, gen.CreateQuizParams{
		CourseID:   quiz.CourseID,
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

func (r *quizRepo) UpdateQuizStatus(ctx context.Context, id uuid.UUID, status domain.QuizStatus) error {
	return r.q.UpdateQuizStatus(ctx, gen.UpdateQuizStatusParams{
		ID:     id,
		Status: string(status),
	})
}

func toDomainQuiz(q gen.Quiz) *domain.Quiz {
	return &domain.Quiz{
		ID:         q.ID,
		CourseID:   q.CourseID,
		Difficulty: domain.Difficulty(q.Difficulty),
		Status:     domain.QuizStatus(q.Status),
		CreatedAt:  q.CreatedAt,
		UpdatedAt:  q.UpdatedAt,
	}
}

// Questions

func (r *quizRepo) CreateQuestions(ctx context.Context, questions []domain.Question) error {
	for _, q := range questions {
		choicesJSON, err := json.Marshal(q.Choices)
		if err != nil {
			return err
		}
		_, err = r.q.CreateQuestion(ctx, gen.CreateQuestionParams{
			QuizID:   q.QuizID,
			Type:     string(q.Type),
			Question: q.Question,
			Choices:  choicesJSON,
			Answer:   q.Answer,
			OrderIdx: int32(q.OrderIdx),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *quizRepo) ListQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error) {
	rows, err := r.q.ListQuestionsByQuiz(ctx, quizID)
	if err != nil {
		return nil, err
	}
	questions := make([]domain.Question, 0, len(rows))
	for _, row := range rows {
		q, err := toDomainQuestion(row)
		if err != nil {
			return nil, err
		}
		questions = append(questions, *q)
	}
	return questions, nil
}

func (r *quizRepo) GetQuestionByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	row, err := r.q.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainQuestion(row)
}
func (r *quizRepo) DeleteQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) error {
	return r.q.DeleteQuestionsByQuiz(ctx, quizID)
}

func toDomainQuestion(q gen.Question) (*domain.Question, error) {
	var choices []domain.Choice
	if len(q.Choices) > 0 {
		if err := json.Unmarshal(q.Choices, &choices); err != nil {
			return nil, err
		}
	}
	return &domain.Question{
		ID:       q.ID,
		QuizID:   q.QuizID,
		Type:     domain.QuestionType(q.Type),
		Question: q.Question,
		Choices:  choices,
		Answer:   q.Answer,
		OrderIdx: int(q.OrderIdx),
	}, nil
}

// Attempts

func (r *quizRepo) CreateAttempt(ctx context.Context, attempt *domain.Attempt) error {
	row, err := r.q.CreateAttempt(ctx, gen.CreateAttemptParams{
		QuizID: attempt.ID,
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
