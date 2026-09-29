package handler

import (
	"net/http"

	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/pkg/apierr"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func (h *AnalyticsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	overview, err := h.svc.GetOverview(r.Context(), claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, overview)
}

func (h *AnalyticsHandler) CourseMetrics(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid course id"))
		return
	}

	metrics, err := h.svc.GetCourseMetrics(r.Context(), courseID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, metrics)
}

func (h *AnalyticsHandler) StudentOverview(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	data, err := h.svc.GetStudentAnalytics(r.Context(), claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, data)
}

func (h *AnalyticsHandler) InstructorOverview(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	data, err := h.svc.GetInstructorAnalytics(r.Context(), claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, data)
}
