package domain

import (
	"time"

	"github.com/google/uuid"
)

type Lesson struct {
	ID        uuid.UUID
	CourseID  uuid.UUID
	Title     string
	OrderIdx  int
	CreatedAt time.Time
	UpdatedAt time.Time
}
