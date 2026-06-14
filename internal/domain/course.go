package domain

import (
"time"

"github.com/google/uuid"
)

type Course struct {
ID           uuid.UUID `json:"id"`
Title        string    `json:"title"`
Description  string    `json:"description"`
Rank         string    `json:"rank"`
InstructorID uuid.UUID `json:"instructor_id"`
Published    bool      `json:"published"`
CreatedAt    time.Time `json:"created_at"`
UpdatedAt    time.Time `json:"updated_at"`
}
