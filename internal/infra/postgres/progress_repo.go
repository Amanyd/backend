package postgres

import (
	"context"
	"time"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/postgres/gen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ProgressRepo struct {
	q *gen.Queries
}

func NewProgressRepo(q *gen.Queries) *ProgressRepo {
	return &ProgressRepo{q: q}
}

func (r *ProgressRepo) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (*domain.UserCourseProgressData, error) {
	result := &domain.UserCourseProgressData{
		CourseID: courseID,
		Lessons:  make(map[uuid.UUID]*domain.LessonProg),
	}

	courseProg, err := r.q.GetCourseProgress(ctx, gen.GetCourseProgressParams{
		UserID:   userID,
		CourseID: courseID,
	})
	if err == nil && len(courseProg) > 0 {
		result.IsCompleted = courseProg[0].IsCompleted
		if courseProg[0].CompletedAt.Valid {
			result.CompletedAt = &courseProg[0].CompletedAt.Time
		}
	}

	lessonProg, _ := r.q.GetLessonProgress(ctx, gen.GetLessonProgressParams{
		UserID:   userID,
		CourseID: courseID,
	})
	
	for _, lp := range lessonProg {
		prog := &domain.LessonProg{
			IsCompleted: lp.IsCompleted,
			ViewedFiles: make(map[uuid.UUID]bool),
		}
		if lp.CompletedAt.Valid {
			prog.CompletedAt = &lp.CompletedAt.Time
		}
		result.Lessons[lp.LessonID] = prog
	}

	fileProg, _ := r.q.GetFileProgress(ctx, gen.GetFileProgressParams{
		UserID:   userID,
		CourseID: courseID,
	})

	for _, fp := range fileProg {
		if result.Lessons[fp.LessonID] == nil {
			result.Lessons[fp.LessonID] = &domain.LessonProg{
				ViewedFiles: make(map[uuid.UUID]bool),
			}
		}
		result.Lessons[fp.LessonID].ViewedFiles[fp.FileID] = fp.IsViewed
	}

	return result, nil
}

func (r *ProgressRepo) UpsertFileProgress(ctx context.Context, userID, fileID uuid.UUID, isViewed bool) error {
	_, err := r.q.UpsertFileProgress(ctx, gen.UpsertFileProgressParams{
		UserID:   userID,
		FileID:   fileID,
		IsViewed: isViewed,
	})
	return err
}

func (r *ProgressRepo) UpsertLessonProgress(ctx context.Context, userID, lessonID uuid.UUID, isCompleted bool) error {
	var completedAt pgtype.Timestamptz
	if isCompleted {
		completedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}
	_, err := r.q.UpsertLessonProgress(ctx, gen.UpsertLessonProgressParams{
		UserID:      userID,
		LessonID:    lessonID,
		IsCompleted: isCompleted,
		CompletedAt: completedAt,
	})
	return err
}

func (r *ProgressRepo) UpsertCourseProgress(ctx context.Context, userID, courseID uuid.UUID, isCompleted bool) error {
	var completedAt pgtype.Timestamptz
	if isCompleted {
		completedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
	}
	_, err := r.q.UpsertCourseProgress(ctx, gen.UpsertCourseProgressParams{
		UserID:      userID,
		CourseID:    courseID,
		IsCompleted: isCompleted,
		CompletedAt: completedAt,
	})
	return err
}
