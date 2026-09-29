package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/pkg/apierr"
	"github.com/Amanyd/backend/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type QuizHandler struct {
	svc *service.QuizService
}

func (h *QuizHandler) ListByCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	claims := GetClaims(r)
	quizzes, err := h.svc.ListByCourseWithUser(r.Context(), courseID, claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, quizzes)
}

func (h *QuizHandler) Get(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(chi.URLParam(r, "quizId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid quiz id"))
		return
	}

	result, err := h.svc.GetQuiz(r.Context(), quizID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, result)
}

func (h *QuizHandler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(chi.URLParam(r, "quizId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid quiz id"))
		return
	}

	claims := GetClaims(r)
	attempt, err := h.svc.StartAttempt(r.Context(), quizID, claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusCreated, attempt)
}

type submitAnswerRequest struct {
	QuestionID string `json:"question_id" validate:"required"`
	Answer     string `json:"answer"      validate:"required"`
}

func (h *QuizHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	attemptID, err := uuid.Parse(chi.URLParam(r, "attemptId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid attempt id"))
		return
	}

	var req submitAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	questionID, err := uuid.Parse(req.QuestionID)
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid question id"))
		return
	}

	answer, err := h.svc.SubmitAnswer(r.Context(), attemptID, questionID, req.Answer)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, answer)
}

func (h *QuizHandler) FinishAttempt(w http.ResponseWriter, r *http.Request) {
	attemptID, err := uuid.Parse(chi.URLParam(r, "attemptId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid attempt id"))
		return
	}

	attempt, err := h.svc.FinishAttempt(r.Context(), attemptID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, attempt)
}

func (h *QuizHandler) Results(w http.ResponseWriter, r *http.Request) {
	attemptID, err := uuid.Parse(chi.URLParam(r, "attemptId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid attempt id"))
		return
	}

	claims := GetClaims(r)
	results, err := h.svc.GetResults(r.Context(), attemptID, claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, results)
}

func (h *QuizHandler) Reset(w http.ResponseWriter, r *http.Request) {
	quizID, err := uuid.Parse(chi.URLParam(r, "quizId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid quiz id"))
		return
	}

	claims := GetClaims(r)
	if err := h.svc.ResetQuiz(r.Context(), quizID, claims.UserID); err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, map[string]string{"status": "generating"})
}
