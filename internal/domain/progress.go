package domain

import (
	"time"

	"github.com/google/uuid"
)

type CourseProgress struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	CourseID    uuid.UUID  `json:"course_id"`
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type LessonProgress struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	LessonID    uuid.UUID  `json:"lesson_id"`
	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type FileProgress struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	FileID   uuid.UUID `json:"file_id"`
	IsViewed bool      `json:"is_viewed"`
}

// Aggregated struct for the frontend
type UserCourseProgressData struct {
	CourseID      uuid.UUID                 `json:"course_id"`
	IsCompleted   bool                      `json:"is_completed"`
	CompletedAt   *time.Time                `json:"completed_at,omitempty"`
	Lessons       map[uuid.UUID]*LessonProg `json:"lessons"`
}

type LessonProg struct {
	IsCompleted bool                    `json:"is_completed"`
	CompletedAt *time.Time              `json:"completed_at,omitempty"`
	ViewedFiles map[uuid.UUID]bool      `json:"viewed_files"`
}
