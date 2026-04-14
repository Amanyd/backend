package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/eulerbutcooler/wingman-backend/internal/infra/postgres/gen"
	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type analyticsRepo struct {
	q *gen.Queries
}

func NewAnalyticsRepo(pool *pgxpool.Pool) port.AnalyticsRepo {
	return &analyticsRepo{q: gen.New(pool)}
}

func (r *analyticsRepo) RecordEvent(ctx context.Context, event *domain.Event) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}

	var courseID pgtype.UUID
	if event.CourseID != nil {
		courseID = pgtype.UUID{Bytes: *event.CourseID, Valid: true}
	}

	row, err := r.q.RecordEvent(ctx, gen.RecordEventParams{
		UserID:   event.UserID,
		CourseID: courseID,
		Type:     string(event.Type),
		Metadata: metadataJSON,
	})
	if err != nil {
		return err
	}
	event.ID = row.ID
	event.CreatedAt = row.CreatedAt
	return nil
}

func (r *analyticsRepo) GetCourseMetrics(ctx context.Context, courseID uuid.UUID) (*domain.Metric, error) {
	row, err := r.q.GetCourseMetrics(ctx, courseID)
	if err != nil {
		return nil, err
	}
	avgScore, err := toFloat64(row.AvgQuizScore)
	if err != nil {
		return nil, fmt.Errorf("avg_quiz_score: %w", err)
	}

	return &domain.Metric{
		CourseID:      row.CourseID,
		TotalStudents: int(row.TotalStudents),
		AvgQuizScore:  avgScore,
		TotalMessages: int(row.TotalMessages),
		TotalFiles:    int(row.TotalFiles),
	}, nil
}

func (r *analyticsRepo) GetStudentScores(ctx context.Context, courseID uuid.UUID) ([]domain.StudentScore, error) {
	rows, err := r.q.GetStudentScores(ctx, courseID)
	if err != nil {
		return nil, err
	}
	scores := make([]domain.StudentScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.StudentScore{
			UserID:   row.UserID,
			Name:     row.Name,
			Rank:     row.Rank,
			AvgScore: row.AvgScore,
		}
	}
	return scores, nil
}

func (r *analyticsRepo) GetOverview(ctx context.Context, instructorID uuid.UUID) (*domain.Overview, error) {
	row, err := r.q.GetOverview(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	avgScore, err := toFloat64(row.AvgScore)
	if err != nil {
		return nil, fmt.Errorf("avg_score: %w", err)
	}
	return &domain.Overview{
		TotalStudents: int(row.TotalStudents),
		TotalCourses:  int(row.TotalCourses),
		AvgScore:      avgScore,
	}, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int64:
		return float64(val), nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}
