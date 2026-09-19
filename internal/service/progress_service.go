package service

import (
	"context"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
)

type ProgressService struct {
	repo port.ProgressRepo
}

func NewProgressService(repo port.ProgressRepo) *ProgressService {
	return &ProgressService{repo: repo}
}

func (s *ProgressService) GetCourseProgress(ctx context.Context, userID, courseID uuid.UUID) (*domain.UserCourseProgressData, error) {
	return s.repo.GetCourseProgress(ctx, userID, courseID)
}

func (s *ProgressService) MarkFileViewed(ctx context.Context, userID, fileID uuid.UUID) error {
	return s.repo.UpsertFileProgress(ctx, userID, fileID, true)
}

func (s *ProgressService) MarkLessonComplete(ctx context.Context, userID, lessonID uuid.UUID) error {
	return s.repo.UpsertLessonProgress(ctx, userID, lessonID, true)
}

func (s *ProgressService) MarkCourseComplete(ctx context.Context, userID, courseID uuid.UUID) error {
	return s.repo.UpsertCourseProgress(ctx, userID, courseID, true)
}
