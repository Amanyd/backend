package domain

import (
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	CourseID  *uuid.UUID  `json:"course_id,omitempty"`
	Title     string      `json:"title"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type Citation struct {
	FileName string   `json:"file_name"`
	FileID   string   `json:"file_id"`
	Score    *float64 `json:"score,omitempty"`
}

type Message struct {
	ID        uuid.UUID   `json:"id"`
	SessionID uuid.UUID   `json:"session_id"`
	Role      MessageRole `json:"role"`
	Content   string      `json:"content"`
	Citations []Citation  `json:"citations,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}
