package postgres

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/eulerbutcooler/wingman-backend/internal/infra/postgres/gen"
	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type lessonRepo struct {
	q *gen.Queries
}

func NewLessonRepo(pool *pgxpool.Pool) port.LessonRepository {
	return &lessonRepo{q: gen.New(pool)}
}

func (r *lessonRepo) Create(ctx context.Context, l *domain.Lesson) error {
	row, err := r.q.CreateLesson(ctx, gen.CreateLessonParams{
		CourseID: l.CourseID,
		Title:    l.Title,
		OrderIdx: int32(l.OrderIdx),
	})
	if err != nil {
		return err
	}
	l.ID = row.ID
	l.CreatedAt = row.CreatedAt
	l.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *lessonRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Lesson, error) {
	row, err := r.q.GetLessonByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainLesson(row), nil
}

func (r *lessonRepo) ListByCourse(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error) {
	rows, err := r.q.ListLessonsByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}
	lessons := make([]domain.Lesson, len(rows))
	for i, row := range rows {
		lessons[i] = *toDomainLesson(row)
	}
	return lessons, nil
}

func (r *lessonRepo) Update(ctx context.Context, l *domain.Lesson) error {
	row, err := r.q.UpdateLesson(ctx, gen.UpdateLessonParams{
		ID:       l.ID,
		Title:    l.Title,
		OrderIdx: int32(l.OrderIdx),
	})
	if err != nil {
		return err
	}
	l.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *lessonRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteLesson(ctx, id)
}

func toDomainLesson(l gen.Lesson) *domain.Lesson {
	return &domain.Lesson{
		ID:        l.ID,
		CourseID:  l.CourseID,
		Title:     l.Title,
		OrderIdx:  int(l.OrderIdx),
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}
