package postgres

import (
	"context"
	"encoding/json"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/postgres/gen"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type chatRepo struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

func NewChatRepo(pool *pgxpool.Pool) port.ChatRepository {
	return &chatRepo{pool: pool, q: gen.New(pool)}
}

func (r *chatRepo) CreateSession(ctx context.Context, s *domain.ChatSession) error {
	var courseID pgtype.UUID
	if s.CourseID != nil {
		courseID = pgtype.UUID{Bytes: *s.CourseID, Valid: true}
	}

	row, err := r.q.CreateChatSession(ctx, gen.CreateChatSessionParams{
		UserID:   s.UserID,
		CourseID: courseID,
		Title:    s.Title,
	})
	if err != nil {
		return err
	}
	s.ID = row.ID
	s.CreatedAt = row.CreatedAt
	return nil
}

func (r *chatRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.ChatSession, error) {
	row, err := r.q.GetChatSessionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDomainSession(row), nil
}

func (r *chatRepo) ListSessionsByUser(ctx context.Context, userID uuid.UUID) ([]domain.ChatSession, error) {
	const query = `
		SELECT 
			s.id, 
			s.user_id, 
			s.course_id, 
			COALESCE(
				NULLIF((SELECT content FROM chat_messages WHERE session_id = s.id AND role = 'user' ORDER BY created_at ASC LIMIT 1), ''),
				s.title
			) AS title,
			s.created_at, 
			s.updated_at
		FROM chat_sessions s
		WHERE s.user_id = $1
		ORDER BY s.updated_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]domain.ChatSession, 0)
	for rows.Next() {
		var s domain.ChatSession
		var courseID *uuid.UUID
		if err := rows.Scan(&s.ID, &s.UserID, &courseID, &s.Title, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.CourseID = courseID
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *chatRepo) CreateMessage(ctx context.Context, m *domain.Message) error {
	citationsJSON, err := json.Marshal(m.Citations)
	if err != nil {
		return err
	}

	row, err := r.q.CreateMessage(ctx, gen.CreateMessageParams{
		SessionID: m.SessionID,
		Role:      string(m.Role),
		Content:   m.Content,
		Citations: citationsJSON,
	})
	if err != nil {
		return err
	}
	m.ID = row.ID
	m.CreatedAt = row.CreatedAt

	// Update session updated_at timestamp so active sessions bubble to top
	_, _ = r.pool.Exec(ctx, "UPDATE chat_sessions SET updated_at = now() WHERE id = $1", m.SessionID)

	return nil
}

func (r *chatRepo) ListMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]domain.Message, error) {
	rows, err := r.q.ListMessagesBySession(ctx, gen.ListMessagesBySessionParams{
		SessionID: sessionID,
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, err
	}
	messages := make([]domain.Message, len(rows))
	for i, row := range rows {
		msg, err := toDomainMessage(row)
		if err != nil {
			return nil, err
		}
		messages[i] = *msg
	}
	return messages, nil
}

func toDomainSession(s gen.ChatSession) *domain.ChatSession {
	var courseID *uuid.UUID
	if s.CourseID.Valid {
		id := uuid.UUID(s.CourseID.Bytes)
		courseID = &id
	}
	return &domain.ChatSession{
		ID:        s.ID,
		UserID:    s.UserID,
		CourseID:  courseID,
		Title:     s.Title,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toDomainMessage(m gen.ChatMessage) (*domain.Message, error) {
	citations := make([]domain.Citation, 0)
	if len(m.Citations) > 0 {
		if err := json.Unmarshal(m.Citations, &citations); err != nil {
			return nil, err
		}
	}
	return &domain.Message{
		ID:        m.ID,
		SessionID: m.SessionID,
		Role:      domain.MessageRole(m.Role),
		Content:   m.Content,
		Citations: citations,
		CreatedAt: m.CreatedAt,
	}, nil
}
