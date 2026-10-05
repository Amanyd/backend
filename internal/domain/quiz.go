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
	ID            uuid.UUID  `json:"id"`
	CourseID      uuid.UUID  `json:"course_id"`
	LessonID      *uuid.UUID `json:"lesson_id"`
	Difficulty    Difficulty `json:"difficulty"`
	Status        QuizStatus `json:"status"`
	QuestionCount int        `json:"question_count"`
	LastScore     *float64   `json:"last_score,omitempty"`
	IsAttempted   bool       `json:"is_attempted"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
	ID          uuid.UUID    `json:"id"`
	QuizID      uuid.UUID    `json:"quiz_id"`
	Type        QuestionType `json:"type"`
	Question    string       `json:"question"`
	Choices     []Choice     `json:"choices"`
	Answer      string       `json:"answer"`
	OrderIdx    int          `json:"order_idx"`
	TopicID     *uuid.UUID   `json:"topic_id,omitempty"`
	Explanation string       `json:"explanation,omitempty"`
	Difficulty  string       `json:"difficulty,omitempty"`
	TopicPhrase string       `json:"topic_phrase,omitempty"`
}

type TopicSlide struct {
	SlideNumber    int      `json:"slide_number"`
	SlideType      string   `json:"slide_type"`
	Title          string   `json:"title"`
	Bullets        []string `json:"bullets"`
	FormulaOrRule  string   `json:"formula_or_rule,omitempty"`
	DiagramMermaid string   `json:"diagram_mermaid,omitempty"`
	Warning        string   `json:"warning,omitempty"`
}

type LessonTopic struct {
	ID         uuid.UUID    `json:"id"`
	LessonID   uuid.UUID    `json:"lesson_id"`
	Title      string       `json:"title"`
	OrderIndex int          `json:"order_index"`
	Slides     []TopicSlide `json:"slides"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
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
