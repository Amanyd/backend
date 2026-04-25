package postgres

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/eulerbutcooler/wingman-backend/internal/infra/postgres/gen"
	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileRepo struct {
	q *gen.Queries
}

func NewFileRepo(pool *pgxpool.Pool) port.FileRepository {
	return &fileRepo{q: gen.New(pool)}
}

func (r *fileRepo) Create(ctx context.Context, f *domain.FileAsset) error {
	row, err := r.q.CreateFile(ctx, gen.CreateFileParams{
		LessonID:     f.LessonID,
		FileName:     f.FileName,
		FileType:     string(f.FileType),
		MinioKey:     f.MinioKey,
		IngestStatus: string(f.IngestStatus),
	})
	if err != nil {
		return err
	}
	f.ID = row.ID
	f.CreatedAt = row.CreatedAt
	return nil
}

func (r *fileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.FileAsset, error) {
	row, err := r.q.GetFileByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainFile(row), nil
}

func (r *fileRepo) ListByLesson(ctx context.Context, lessonID uuid.UUID) ([]domain.FileAsset, error) {
	rows, err := r.q.ListFilesByLesson(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	files := make([]domain.FileAsset, len(rows))
	for i, row := range rows {
		files[i] = *toDomainFile(row)
	}
	return files, nil
}

func (r *fileRepo) UpdateIngestStatus(ctx context.Context, fileID uuid.UUID, status domain.IngestStatus) error {
	return r.q.UpdateFileIngestStatus(ctx, gen.UpdateFileIngestStatusParams{
		ID:           fileID,
		IngestStatus: string(status),
	})
}

func (r *fileRepo) AllReadyForCourse(ctx context.Context, courseID uuid.UUID) (bool, error) {
	result, err := r.q.AllFilesReadyForCourse(ctx, courseID)
	if err != nil {
		return false, err
	}
	return result, nil
}

func toDomainFile(f gen.File) *domain.FileAsset {
	return &domain.FileAsset{
		ID:           f.ID,
		LessonID:     f.LessonID,
		FileName:     f.FileName,
		FileType:     domain.FileType(f.FileType),
		MinioKey:     f.MinioKey,
		IngestStatus: domain.IngestStatus(f.IngestStatus),
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}
