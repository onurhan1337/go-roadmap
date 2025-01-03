package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/router"
)

type TransactionHandler struct {
	transactionService models.TransactionService
	logger            *logger.Logger
}

func NewTransactionHandler(transactionService models.TransactionService, logger *logger.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:            logger,
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
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Credit(r.Context(), userID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process credit", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleDebit(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Debit(r.Context(), userID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process debit", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	fromUserID := auth.GetUserIDFromContext(r.Context())
	if fromUserID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionService.Transfer(r.Context(), fromUserID, req.ToUserID, req.Amount, req.Notes); err != nil {
		h.logger.Error("failed to process transfer", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *TransactionHandler) HandleGetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	params := r.Context().Value(router.PathParamsKey).(map[string]string)
	transIDStr := params["id"]
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
		h.logger.Error("failed to get transaction", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}