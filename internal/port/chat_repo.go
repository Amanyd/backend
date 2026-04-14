package port

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/domain"
	"github.com/google/uuid"
)

type ChatRepository interface {
	CreateSession(ctx context.Context, session *domain.ChatSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.ChatSession, error)
	ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ChatSession, error)
	CreateMessage(ctx context.Context, msg *domain.Message) error
	ListMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]domain.Message, error)
}
