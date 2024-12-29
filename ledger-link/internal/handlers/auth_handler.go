package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"ledger-link/internal/models"
	"ledger-link/internal/services"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/validator"
)

var (
	ErrInvalidToken = errors.New("invalid token")
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

	token, err := h.authSvc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if err == models.ErrInvalidCredentials {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error("failed to login user", "error", err)
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	oldToken := r.Header.Get("Authorization")
	if oldToken == "" {
		http.Error(w, "Authorization token is required", http.StatusBadRequest)
		return
	}

	if len(oldToken) > 7 && oldToken[:7] == "Bearer " {
		oldToken = oldToken[7:]
	}

	newToken, err := h.authSvc.RefreshToken(r.Context(), oldToken)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": newToken})
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

	token, err := h.authSvc.Register(r.Context(), input.Email, input.Password, input.Username)
	if err != nil {
		if err == models.ErrEmailAlreadyExists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		h.logger.Error("failed to register user", "error", err)
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}