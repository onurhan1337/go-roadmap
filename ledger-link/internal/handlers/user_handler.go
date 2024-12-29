package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledger-link/internal/models"
	"ledger-link/internal/services"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/router"
)

type UserHandler struct {
	userSvc *services.UserService
	logger  *logger.Logger
}

func NewUserHandler(userSvc *services.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		logger:  logger,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userSvc.GetUsers(r.Context())
	if err != nil {
		h.logger.Error("failed to get users", "error", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userSvc.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get user", "error", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := &models.User{
		ID:       uint(id),
		Username: input.Username,
		Email:    input.Email,
	}

	if err := h.userSvc.Update(r.Context(), user); err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to update user", "error", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := router.GetParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userSvc.Delete(r.Context(), uint(id)); err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to delete user", "error", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}