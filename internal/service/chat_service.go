package service

import (
	"context"
	"io"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
)

const chatHistoryLimit = 10

type ChatService struct {
	chats   port.ChatRepository
	courses port.CourseRepository
	users   port.UserRepository
	rag     port.RagClient
}

func NewChatService(
	chats port.ChatRepository,
	courses port.CourseRepository,
	users port.UserRepository,
	rag port.RagClient,
) *ChatService {
	return &ChatService{chats: chats, courses: courses, users: users, rag: rag}
}

func (s *ChatService) CreateSession(ctx context.Context, userID uuid.UUID, courseID *uuid.UUID) (*domain.ChatSession, error) {
	session := &domain.ChatSession{
		UserID:   userID,
		CourseID: courseID,
		Title:    "New Chat",
	}
	if err := s.chats.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *ChatService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.ChatSession, error) {
	return s.chats.ListSessionsByUser(ctx, userID)
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID uuid.UUID, query string) (io.ReadCloser, error) {
	session, err := s.chats.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	courseIDs, err := s.resolveCourseIDs(ctx, session, userID)
	if err != nil {
		return nil, err
	}

	messages, err := s.chats.ListMessages(ctx, sessionID, chatHistoryLimit)
	if err != nil {
		return nil, err
	}

	userMsg := &domain.Message{
		SessionID: sessionID,
		Role:      domain.RoleUser,
		Content:   query,
	}
	if err := s.chats.CreateMessage(ctx, userMsg); err != nil {
		return nil, err
	}

	history := make([]port.ChatMessage, len(messages))
	for i, m := range messages {
		history[i] = port.ChatMessage{Role: string(m.Role), Content: m.Content}
	}

	ids := make([]string, len(courseIDs))
	for i, id := range courseIDs {
		ids[i] = id.String()
	}

	stream, err := s.rag.ChatStream(ctx, port.ChatRequest{
		CourseIDs: ids,
		Query:    query,
		History:  history,
	})
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (s *ChatService) SaveAssistantMessage(ctx context.Context, sessionID uuid.UUID, content string, citations []domain.Citation) error {
	assistantMsg := &domain.Message{
		SessionID: sessionID,
		Role:      domain.RoleAssistant,
		Content:   content,
		Citations: citations,
	}
	return s.chats.CreateMessage(ctx, assistantMsg)
}

func (s *ChatService) GetHistory(ctx context.Context, sessionID uuid.UUID) ([]domain.Message, error) {
	return s.chats.ListMessages(ctx, sessionID, 100)
}

func (s *ChatService) resolveCourseIDs(ctx context.Context, session *domain.ChatSession, userID uuid.UUID) ([]uuid.UUID, error) {
	if session.CourseID != nil {
		return []uuid.UUID{*session.CourseID}, nil
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	courses, err := s.courses.ListByRank(ctx, user.Rank)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(courses))
	for i, c := range courses {
		ids[i] = c.ID
	}
	return ids, nil
}
