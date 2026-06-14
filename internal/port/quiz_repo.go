package port

import (
	"context"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/google/uuid"
)

type QuizRepository interface {
	CreateQuiz(ctx context.Context, quiz *domain.Quiz) error
	GetQuizByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error)
	GetQuizByCourseAndDifficulty(ctx context.Context, courseID uuid.UUID, difficulty domain.Difficulty) (*domain.Quiz, error)
	ListQuizzesByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Quiz, error)
	UpdateQuizStatus(ctx context.Context, id uuid.UUID, status domain.QuizStatus) error
	DeleteQuizzesByCourse(ctx context.Context, courseID uuid.UUID) error

	CreateQuestions(ctx context.Context, questions []domain.Question) error
	ListQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) ([]domain.Question, error)
	GetQuestionByID(ctx context.Context, id uuid.UUID) (*domain.Question, error)
	DeleteQuestionsByQuiz(ctx context.Context, quizID uuid.UUID) error

	CreateAttempt(ctx context.Context, attempt *domain.Attempt) error
	GetAttemptByID(ctx context.Context, id uuid.UUID) (*domain.Attempt, error)
	UpdateAttempt(ctx context.Context, attempt *domain.Attempt) error

	CreateAnswer(ctx context.Context, answer *domain.Answer) error
	ListAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]domain.Answer, error)
}
