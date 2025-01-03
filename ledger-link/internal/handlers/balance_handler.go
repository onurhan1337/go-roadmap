package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type BalanceHandler struct {
	balanceService models.BalanceService
	logger         *logger.Logger
	config         *BalanceHandlerConfig
}

func NewBalanceHandler(balanceService models.BalanceService, logger *logger.Logger, config *BalanceHandlerConfig) *BalanceHandler {
	if config == nil {
		config = DefaultBalanceHandlerConfig()
	}
	return &BalanceHandler{
		balanceService: balanceService,
		logger:         logger,
		config:         config,
	}
}

// GetCurrentBalance returns the current user's balance
func (h *BalanceHandler) GetCurrentBalance(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), user.ID)
	if err != nil {
		h.logger.Error("failed to get balance", "error", err, "user_id", user.ID)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetBalanceHistory returns the current user's balance history
func (h *BalanceHandler) GetBalanceHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := h.config.DefaultHistoryLimit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > h.config.MaxHistoryLimit {
				limit = h.config.MaxHistoryLimit
			} else {
				limit = parsedLimit
			}
		}
	}

	history, err := h.balanceService.GetBalanceHistory(r.Context(), user.ID, limit)
	if err != nil {
		h.logger.Error("failed to get balance history", "error", err, "user_id", user.ID)
		http.Error(w, "Failed to get balance history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
