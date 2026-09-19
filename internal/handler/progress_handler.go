package handler

import (
	"net/http"

	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/pkg/apierr"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProgressHandler struct {
	svc *service.ProgressService
}

func NewProgressHandler(svc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) GetCourseProgress(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	claims := GetClaims(r)
	progress, err := h.svc.GetCourseProgress(r.Context(), claims.UserID, courseID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}

	apierr.WriteData(w, http.StatusOK, progress)
}

func (h *ProgressHandler) MarkFileViewed(w http.ResponseWriter, r *http.Request) {
	fileID, err := uuid.Parse(chi.URLParam(r, "fileId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid file id"))
		return
	}

	claims := GetClaims(r)
	if err := h.svc.MarkFileViewed(r.Context(), claims.UserID, fileID); err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}

	apierr.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProgressHandler) MarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid lesson id"))
		return
	}

	claims := GetClaims(r)
	if err := h.svc.MarkLessonComplete(r.Context(), claims.UserID, lessonID); err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}

	apierr.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *ProgressHandler) MarkCourseComplete(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	claims := GetClaims(r)
	if err := h.svc.MarkCourseComplete(r.Context(), claims.UserID, courseID); err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}

	apierr.WriteData(w, http.StatusOK, map[string]string{"status": "ok"})
}
