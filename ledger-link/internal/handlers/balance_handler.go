package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

func (h *BalanceHandler) GetCurrentBalance(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get balance", "error", err, "user_id", userID)
		http.Error(w, "Failed to get balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) GetHistoricalBalances(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
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

	history, err := h.balanceService.GetBalanceHistory(r.Context(), userID, limit)
	if err != nil {
		h.logger.Error("failed to get balance history", "error", err, "user_id", userID)
		http.Error(w, "Failed to get balance history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *BalanceHandler) GetBalanceAtTime(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	timestampStr := r.URL.Query().Get("timestamp")
	if timestampStr == "" {
		http.Error(w, "timestamp parameter is required", http.StatusBadRequest)
		return
	}

	timestamp, err := time.Parse(h.config.TimeFormat, timestampStr)
	if err != nil {
		http.Error(w, "invalid timestamp format, use "+h.config.TimeFormat, http.StatusBadRequest)
		return
	}

	balance, err := h.balanceService.GetBalanceAtTime(r.Context(), userID, timestamp)
	if err != nil {
		h.logger.Error("failed to get balance at time", "error", err, "user_id", userID, "timestamp", timestamp)
		http.Error(w, "Failed to get balance at time", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}
