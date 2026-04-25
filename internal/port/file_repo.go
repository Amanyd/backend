package port

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/google/uuid"
)

type FileRepository interface {
	Create(ctx context.Context, file *domain.FileAsset) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.FileAsset, error)
	ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]domain.FileAsset, error)
	UpdateIngestStatus(ctx context.Context, fileID uuid.UUID, status domain.IngestStatus) error
	AllReadyForCourse(ctx context.Context, courseID uuid.UUID) (bool, error)
}
