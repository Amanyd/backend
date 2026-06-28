package port

import (
	"context"
	"io"
)

type RagClient interface {
	ChatStream(ctx context.Context, req ChatRequest) (io.ReadCloser, error)
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	GradeAnswer(ctx context.Context, req GradeRequest) (*GradeResponse, error)
}

type ChatRequest struct {
	CourseIDs []string      `json:"course_ids"`
	Query     string        `json:"query"`
	Stream    bool          `json:"stream"`
	History   []ChatMessage `json:"history"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Answer    string         `json:"answer"`
	Citations []CitationItem `json:"citations"`
}

type CitationItem struct {
	FileName string   `json:"file_name"`
	FileID   string   `json:"file_id"`
	Score    *float64 `json:"score,omitempty"`
}

type GradeRequest struct {
	Question        string `json:"question"`
	ReferenceAnswer string `json:"reference_answer"`
	UserAnswer      string `json:"user_answer"`
}

type GradeResponse struct {
	IsCorrect   bool    `json:"is_correct"`
	Score       float64 `json:"score"`
	Explanation string  `json:"explanation"`
}
