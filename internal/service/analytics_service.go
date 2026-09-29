package service

import (
	"context"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
)

type AnalyticsService struct {
	analytics port.AnalyticsRepo
}

func NewAnalyticsService(analytics port.AnalyticsRepo) *AnalyticsService {
	return &AnalyticsService{analytics: analytics}
}

func (s *AnalyticsService) RecordEvent(ctx context.Context, userID uuid.UUID, courseID *uuid.UUID, eventType domain.EventType, metadata map[string]any) error {
	event := &domain.Event{
		UserID:   userID,
		CourseID: courseID,
		Type:     eventType,
		Metadata: metadata,
	}
	return s.analytics.RecordEvent(ctx, event)
}

func (s *AnalyticsService) GetCourseMetrics(ctx context.Context, courseID uuid.UUID) (*domain.Metric, error) {
	return s.analytics.GetCourseMetrics(ctx, courseID)
}

func (s *AnalyticsService) GetStudentScores(ctx context.Context, courseID uuid.UUID) ([]domain.StudentScore, error) {
	return s.analytics.GetStudentScores(ctx, courseID)
}

func (s *AnalyticsService) GetOverview(ctx context.Context, instructorID uuid.UUID) (*domain.Overview, error) {
	return s.analytics.GetOverview(ctx, instructorID)
}

func (s *AnalyticsService) GetStudentAnalytics(ctx context.Context, userID uuid.UUID) (*domain.StudentAnalytics, error) {
	return s.analytics.GetStudentAnalytics(ctx, userID)
}

func (s *AnalyticsService) GetInstructorAnalytics(ctx context.Context, instructorID uuid.UUID) (*domain.InstructorAnalytics, error) {
	return s.analytics.GetInstructorAnalytics(ctx, instructorID)
}
