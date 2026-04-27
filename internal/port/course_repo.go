package port

import (
	"context"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/google/uuid"
)

type CourseRepository interface {
        Create(ctx context.Context, course *domain.Course) error
        GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error)
        ListAll(ctx context.Context) ([]domain.Course, error)
        ListByRank(ctx context.Context, rank string) ([]domain.Course, error)
        ListByInstructor(ctx context.Context, instructorID uuid.UUID) ([]domain.Course, error)
        Update(ctx context.Context, course *domain.Course) error
        Delete(ctx context.Context, id uuid.UUID) error
}