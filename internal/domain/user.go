package domain

import (
"time"

"github.com/google/uuid"
)

type Role string

const (
RoleStudent    Role = "student"
RoleInstructor Role = "instructor"
)

type User struct {
ID           uuid.UUID `json:"id"`
Name         string    `json:"name"`
EnrollmentID string    `json:"enrollment_id"`
Rank         string    `json:"rank"`
Role         Role      `json:"role"`
PasswordHash string    `json:"-"`
CreatedAt    time.Time `json:"created_at"`
UpdatedAt    time.Time `json:"updated_at"`
}
