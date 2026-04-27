package postgres

import (
	"context"
	"errors"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/postgres/gen"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type courseRepo struct {
        q    *gen.Queries
        pool *pgxpool.Pool
}

func NewCourseRepo(pool *pgxpool.Pool) port.CourseRepository {
        return &courseRepo{q: gen.New(pool), pool: pool}
}

func (r *courseRepo) Create(ctx context.Context, course *domain.Course) error {
	row, err := r.q.CreateCourse(ctx, gen.CreateCourseParams{
		Title:        course.Title,
		Description:  course.Description,
		Rank:         course.Rank,
		InstructorID: course.InstructorID,
	})
	if err != nil {
		return err
	}
	course.ID = row.ID
	course.CreatedAt = row.CreatedAt
	course.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *courseRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Course, error) {
	row, err := r.q.GetCourseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomainCourse(row), nil
}

func (r *courseRepo) ListAll(ctx context.Context) ([]domain.Course, error) {
        rows, err := r.pool.Query(ctx, "SELECT id, title, description, rank, instructor_id, created_at, updated_at FROM courses ORDER BY created_at DESC")
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var courses []domain.Course
        for rows.Next() {
                var c gen.Course
                if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Rank, &c.InstructorID, &c.CreatedAt, &c.UpdatedAt); err != nil {
                        return nil, err
                }
                courses = append(courses, *toDomainCourse(c))
        }
        return courses, rows.Err()
}

func (r *courseRepo) ListByRank(ctx context.Context, rank string) ([]domain.Course, error) {
	rows, err := r.q.ListCoursesByRank(ctx, rank)
	if err != nil {
		return nil, err
	}
	courses := make([]domain.Course, len(rows))
	for i, row := range rows {
		courses[i] = *toDomainCourse(row)
	}
	return courses, nil
}

func (r *courseRepo) ListByInstructor(ctx context.Context, instructorID uuid.UUID) ([]domain.Course, error) {
	rows, err := r.q.ListCoursesByInstructor(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	courses := make([]domain.Course, len(rows))
	for i, row := range rows {
		courses[i] = *toDomainCourse(row)
	}
	return courses, nil
}

func (r *courseRepo) Update(ctx context.Context, course *domain.Course) error {
	row, err := r.q.UpdateCourse(ctx, gen.UpdateCourseParams{
		ID:          course.ID,
		Title:       course.Title,
		Description: course.Description,
		Rank:        course.Rank,
	})
	if err != nil {
		return err
	}
	course.UpdatedAt = row.UpdatedAt
	return nil
}
func (r *courseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteCourse(ctx, id)
}
func toDomainCourse(course gen.Course) *domain.Course {
	return &domain.Course{
		ID:           course.ID,
		Title:        course.Title,
		Description:  course.Description,
		Rank:         course.Rank,
		InstructorID: course.InstructorID,
		CreatedAt:    course.CreatedAt,
		UpdatedAt:    course.UpdatedAt,
	}
}
