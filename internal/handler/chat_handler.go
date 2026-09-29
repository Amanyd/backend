package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/pkg/apierr"
	"github.com/Amanyd/backend/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ChatHandler struct {
	svc *service.ChatService
}

type createSessionRequest struct {
	CourseID *string `json:"course_id"`
}

func (h *ChatHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}

	claims := GetClaims(r)

	var courseID *uuid.UUID
	if req.CourseID != nil {
		id, err := uuid.Parse(*req.CourseID)
		if err != nil {
			apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
			return
		}
		courseID = &id
	}

	session, err := h.svc.CreateSession(r.Context(), claims.UserID, courseID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusCreated, session)
}

func (h *ChatHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	sessions, err := h.svc.ListSessions(r.Context(), claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, sessions)
}

type sendMessageRequest struct {
	Query string `json:"query" validate:"required"`
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid session id"))
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	claims := GetClaims(r)
	stream, err := h.svc.SendMessage(r.Context(), sessionID, claims.UserID, req.Query)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	defer stream.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		apierr.WriteJSON(w, apierr.Internal("streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Flush headers immediately so the client knows the connection is alive.
	// Without this, the upstream proxy (Next.js) will time out waiting for
	// the first SSE token while the LLM is still generating.
	flusher.Flush()

	var assistantContent strings.Builder
	var citations []domain.Citation

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()

		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if payload != "[DONE]" && payload != "" {
				var frame struct {
					Token     string            `json:"token"`
					Citations []domain.Citation `json:"citations"`
				}
				if err := json.Unmarshal([]byte(payload), &frame); err == nil {
					if frame.Token != "" {
						assistantContent.WriteString(frame.Token)
					}
					if len(frame.Citations) > 0 {
						citations = frame.Citations
					}
				}
			}
		}
	}

	if assistantContent.Len() > 0 {
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = h.svc.SaveAssistantMessage(saveCtx, sessionID, assistantContent.String(), citations)
	}
}

func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid session id"))
		return
	}

	messages, err := h.svc.GetHistory(r.Context(), sessionID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, messages)
}
