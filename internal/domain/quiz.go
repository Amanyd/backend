package domain

import (
	"time"

	"github.com/google/uuid"
)

type QuizStatus string

const (
	QuizPending    QuizStatus = "pending"
	QuizGenerating QuizStatus = "generating"
	QuizReady      QuizStatus = "ready"
	QuizFailed     QuizStatus = "failed"
)

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type Quiz struct {
	ID         uuid.UUID  `json:"id"`
	CourseID   uuid.UUID  `json:"course_id"`
	Difficulty Difficulty `json:"difficulty"`
	Status     QuizStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type QuestionType string

const (
	QuestionMCQ       QuestionType = "mcq"
	QuestionOpenEnded QuestionType = "open_ended"
)

type Choice struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type Question struct {
	ID       uuid.UUID    `json:"id"`
	QuizID   uuid.UUID    `json:"quiz_id"`
	Type     QuestionType `json:"type"`
	Question string       `json:"question"`
	Choices  []Choice      `json:"choices"`
	Answer   string       `json:"answer"`
	OrderIdx int          `json:"order_idx"`
}

type Attempt struct {
	ID        uuid.UUID  `json:"id"`
	QuizID    uuid.UUID  `json:"quiz_id"`
	UserID    uuid.UUID  `json:"user_id"`
	Score     float64    `json:"score"`
	Total     int        `json:"total"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type Answer struct {
	ID         uuid.UUID `json:"id"`
	AttemptID  uuid.UUID `json:"attempt_id"`
	QuestionID uuid.UUID `json:"question_id"`
	UserAnswer string    `json:"user_answer"`
	IsCorrect  bool      `json:"is_correct"`
}
