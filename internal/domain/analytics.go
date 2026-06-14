package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

type Event struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	CourseID  *uuid.UUID     `json:"course_id,omitempty"`
	Type      EventType      `json:"type"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Metric struct {
	CourseID      uuid.UUID `json:"course_id"`
	TotalStudents int       `json:"total_students"`
	AvgQuizScore  float64   `json:"avg_quiz_score"`
	TotalMessages int       `json:"total_messages"`
	TotalFiles    int       `json:"total_files"`
}

type StudentScore struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Rank     string    `json:"rank"`
	AvgScore float64   `json:"avg_score"`
}

type Overview struct {
	TotalStudents int     `json:"total_students"`
	TotalCourses  int     `json:"total_courses"`
	AvgScore      float64 `json:"avg_score"`
}
