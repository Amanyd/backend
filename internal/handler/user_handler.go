package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/pkg/apierr"
	"github.com/Amanyd/backend/pkg/logger"
	"github.com/Amanyd/backend/pkg/validator"
	"go.uber.org/zap"
)

type UserHandler struct {
	svc *service.UserService
}

type registerRequest struct {
	Name         string `json:"name"          validate:"required,min=2"`
	EnrollmentID string `json:"enrollment_id" validate:"required"`
	Rank         string `json:"rank"          validate:"required,rank"`
	Password     string `json:"password"      validate:"required,min=6"`
	Role         string `json:"role"          validate:"required,oneof=student instructor"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Name, req.EnrollmentID, req.Rank, req.Password, domain.Role(req.Role))
	if err != nil {
		logger.FromCtx(r.Context()).Error("register failed", zap.Error(err), zap.String("enrollment_id", req.EnrollmentID))
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusCreated, user)
}

type loginRequest struct {
	EnrollmentID string `json:"enrollment_id" validate:"required"`
	Password     string `json:"password"      validate:"required"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	tokens, err := h.svc.Login(r.Context(), req.EnrollmentID, req.Password)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, tokens)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (h *UserHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierr.WriteJSON(w, apierr.BadRequest("invalid json"))
		return
	}
	if err := validator.ValidateStruct(req); err != nil {
		apierr.WriteJSON(w, err)
		return
	}

	accessToken, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, map[string]string{"access_token": accessToken})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := GetClaims(r)
	user, err := h.svc.GetProfile(r.Context(), claims.UserID)
	if err != nil {
		apierr.WriteJSON(w, mapDomainError(err))
		return
	}
	apierr.WriteData(w, http.StatusOK, user)
}
