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

type LessonHandler struct {
	svc *service.CourseService
}

type createLessonRequest struct {
	Title    string `json:"title"     validate:"required,min=2"`
	OrderIdx int    `json:"order_idx"`
}

func (h *LessonHandler) Create(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	var req createLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	claims := GetClaims(r)
	lesson, err := h.svc.CreateLesson(r.Context(), courseID, claims.UserID, req.Title, req.OrderIdx)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusCreated, lesson)
}

func (h *LessonHandler) List(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	lessons, err := h.svc.ListLessons(r.Context(), courseID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, lessons)
}

type updateLessonRequest struct {
	Title    string `json:"title"     validate:"required,min=2"`
	OrderIdx int    `json:"order_idx"`
}

func (h *LessonHandler) Update(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid lesson id"))
		return
	}

	var req updateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	claims := GetClaims(r)
	lesson, err := h.svc.UpdateLesson(r.Context(), lessonID, claims.UserID, req.Title, req.OrderIdx)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, lesson)
}

func (h *LessonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid lesson id"))
		return
	}

	claims := GetClaims(r)
	if err := h.svc.DeleteLesson(r.Context(), lessonID, claims.UserID); err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *LessonHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid lesson id"))
		return
	}

	topics, err := h.svc.ListLessonTopics(r.Context(), lessonID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, topics)
}
