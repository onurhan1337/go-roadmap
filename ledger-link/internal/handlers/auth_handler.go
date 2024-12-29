package handlers

import (
	"encoding/json"
	"net/http"

	"ledger-link/internal/models"
	"ledger-link/internal/services"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/validator"
)

type AuthHandler struct {
	authSvc *services.AuthService
	logger  *logger.Logger
}

func NewAuthHandler(authSvc *services.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authSvc: authSvc,
		logger:  logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validator.Validate(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.authSvc.Register(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to register user", "error", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validator.Validate(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.authSvc.Login(r.Context(), input)
	if err != nil {
		if err == models.ErrInvalidCredentials || err == services.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to login user", "error", err)
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("X-Refresh-Token")
	if refreshToken == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	response, err := h.authSvc.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		if err == models.ErrInvalidToken {
			http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to refresh token", "error", err)
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}