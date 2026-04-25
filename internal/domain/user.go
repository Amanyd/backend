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
	ID           uuid.UUID
	Name         string
	EnrollmentID string
	Rank         string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
