package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/httputil"
	"ledger-link/pkg/logger"
)

type TransactionHandler struct {
	transactionService models.TransactionService
	logger             *logger.Logger
}

func NewTransactionHandler(transactionService models.TransactionService, logger *logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

type TransactionRequest struct {
	Amount float64 `json:"amount"`
	Notes  string  `json:"notes"`
}

type TransferRequest struct {
	ToUserID uint    `json:"to_user_id"`
	Amount   float64 `json:"amount"`
	Notes    string  `json:"notes"`
}

func (h *TransactionHandler) HandleCredit(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Credit(r.Context(), user.ID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process credit", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleDebit(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Debit(r.Context(), user.ID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process debit", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Transfer(r.Context(), user.ID, req.ToUserID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process transfer", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleGetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var userID uint
	if user.Role == models.RoleAdmin {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr != "" {
			id, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				http.Error(w, "invalid user ID", http.StatusBadRequest)
				return
			}
			userID = uint(id)
		}
	} else {
		userID = user.ID
	}

	if userID == 0 {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	transactions, err := h.transactionService.GetUserTransactions(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get transaction history", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) HandleGetTransaction(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	transIDStr := httputil.GetPathParam(r.Context(), "id")
	if transIDStr == "" {
		http.Error(w, "transaction ID is required", http.StatusBadRequest)
		return
	}

	transID, err := strconv.ParseUint(transIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction ID", http.StatusBadRequest)
		return
	}

	transaction, err := h.transactionService.GetTransaction(r.Context(), uint(transID))
	if err != nil {
		if err == models.ErrNotFound {
			http.Error(w, "transaction not found", http.StatusNotFound)
			return
		}
		h.logger.Error("failed to get transaction", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if user.Role != models.RoleAdmin && transaction.FromUserID != user.ID && transaction.ToUserID != user.ID {
		h.logger.Error("unauthorized access to transaction",
			"user_id", user.ID,
			"transaction_from", transaction.FromUserID,
			"transaction_to", transaction.ToUserID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}
