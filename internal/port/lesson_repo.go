package port

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/google/uuid"
)

type LessonRepository interface {
	Create(ctx context.Context, lesson *domain.Lesson) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error)
	ListByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error)
	Update(ctx context.Context, lesson *domain.Lesson) error
	Delete(ctx context.Context, id uuid.UUID) error
}
