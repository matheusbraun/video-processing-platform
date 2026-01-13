package controller

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/video-platform/services/auth/internal/controller"
	"github.com/video-platform/services/auth/internal/infrastructure/api/dto"
	"github.com/video-platform/services/auth/internal/presenter"
	"github.com/video-platform/services/auth/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/rest"
)

type AuthHTTPController struct {
	controller controller.AuthController
	presenter  presenter.AuthPresenter
}

func NewAuthHTTPController(
	controller controller.AuthController,
	presenter presenter.AuthPresenter,
) *AuthHTTPController {
	return &AuthHTTPController{
		controller: controller,
		presenter:  presenter,
	}
}

func (h *AuthHTTPController) RegisterRoutes(r chi.Router) {
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
}

func (h *AuthHTTPController) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	cmd := commands.RegisterCommand{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.controller.Register(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentRegister(output)
	rest.RespondCreated(w, response)
}

func (h *AuthHTTPController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	cmd := commands.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	output, err := h.controller.Login(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentLogin(output)
	rest.RespondSuccess(w, response)
}

func (h *AuthHTTPController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	cmd := commands.RefreshCommand{
		RefreshToken: req.RefreshToken,
	}

	output, err := h.controller.Refresh(r.Context(), cmd)
	if err != nil {
		rest.RespondError(w, http.StatusUnauthorized, "REFRESH_FAILED", err.Error())
		return
	}

	response := h.presenter.PresentRefresh(output)
	rest.RespondSuccess(w, response)
}

func (h *AuthHTTPController) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body")
		return
	}

	cmd := commands.LogoutCommand{
		RefreshToken: req.RefreshToken,
	}

	if err := h.controller.Logout(r.Context(), cmd); err != nil {
		rest.RespondError(w, http.StatusBadRequest, "LOGOUT_FAILED", err.Error())
		return
	}

	rest.RespondSuccess(w, map[string]string{"message": "logged out successfully"})
}
