package port

import (
	"context"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/google/uuid"
)

type ProgressRepo interface {
	GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (*domain.UserCourseProgressData, error)
	UpsertFileProgress(ctx context.Context, userID, fileID uuid.UUID, isViewed bool) error
	UpsertLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, isCompleted bool) error
	UpsertCourseProgress(ctx context.Context, userID, courseID uuid.UUID, isCompleted bool) error
}
