package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Lesson struct {
	ID        uuid.UUID       `json:"id"`
	CourseID  uuid.UUID       `json:"course_id"`
	Title     string          `json:"title"`
	OrderIdx  int             `json:"order_idx"`
	Keywords  json.RawMessage `json:"keywords,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
