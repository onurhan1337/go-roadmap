package services

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shopspring/decimal"

	"ledger-link/internal/models"
	"ledger-link/internal/processor"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

var (
	transactionCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_transactions_total",
			Help: "The total number of processed transactions by type",
		},
		[]string{"type", "status"},
	)

	transactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_transaction_duration_seconds",
			Help:    "Transaction processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	balanceGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ledger_user_balance",
			Help: "Current balance for users",
		},
		[]string{"user_id"},
	)

	transactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_transaction_amount",
			Help:    "Distribution of transaction amounts",
			Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000, 10000},
		},
		[]string{"type"},
	)

	transactionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_transaction_errors_total",
			Help: "Total number of transaction errors by type and error kind",
		},
		[]string{"type", "error"},
	)

	activeTransactions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ledger_active_transactions",
			Help: "Number of currently processing transactions",
		},
		[]string{"type"},
	)
)

type TransactionService struct {
	repo       models.TransactionRepository
	processor  *processor.TransactionProcessor
	balanceSvc models.BalanceService
	auditSvc   models.AuditService
	logger     *logger.Logger
}

func NewTransactionService(
	repo models.TransactionRepository,
	balanceSvc models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *TransactionService {
	return &TransactionService{
		repo:       repo,
		balanceSvc: balanceSvc,
		auditSvc:   auditSvc,
		logger:     logger,
		processor:  processor.NewTransactionProcessor(repo, balanceSvc, auditSvc, logger),
	}
}

func (s *TransactionService) Credit(ctx context.Context, userID uint, amount decimal.Decimal, notes string) error {
	timer := prometheus.NewTimer(transactionDuration.WithLabelValues("credit"))
	defer timer.ObserveDuration()

	if amount.IsNegative() || amount.IsZero() {
		transactionErrors.WithLabelValues("credit", "invalid_amount").Inc()
		return models.ErrInvalidAmount
	}

	tx := &models.Transaction{
		ToUserID:   userID,
		FromUserID: userID,
		Amount:     amount,
		Type:       models.TypeDeposit,
		Status:     models.StatusPending,
		Notes:      notes,
	}

	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	if err := s.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	if err := s.processor.SubmitForBatchProcessing(tx); err != nil {
		return fmt.Errorf("failed to submit transaction for processing: %w", err)
	}

	transactionCounter.WithLabelValues("credit", "pending").Inc()
	return nil
}

func (s *TransactionService) Debit(ctx context.Context, userID uint, amount decimal.Decimal, notes string) error {
	timer := prometheus.NewTimer(transactionDuration.WithLabelValues("debit"))
	defer timer.ObserveDuration()

	if amount.IsNegative() || amount.IsZero() {
		return models.ErrInvalidAmount
	}

	tx := &models.Transaction{
		FromUserID: userID,
		ToUserID:   userID,
		Amount:     amount,
		Type:       models.TypeWithdrawal,
		Status:     models.StatusPending,
		Notes:      notes,
	}

	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	balance, err := s.balanceSvc.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.SafeAmount().LessThan(amount) {
		return fmt.Errorf("insufficient funds")
	}

	if err := s.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	if err := balance.SubtractAmount(amount); err != nil {
		tx.Status = models.StatusFailed
		if updateErr := s.ProcessTransaction(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction", "error", updateErr)
		}
		return fmt.Errorf("failed to debit amount: %w", err)
	}

	if err := s.balanceSvc.UpdateBalance(ctx, userID, balance.SafeAmount()); err != nil {
		if rbErr := balance.AddAmount(amount); rbErr != nil {
			s.logger.Error("failed to rollback balance update", "error", rbErr)
		}
		tx.Status = models.StatusFailed
		if updateErr := s.ProcessTransaction(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction", "error", updateErr)
		}
		return fmt.Errorf("failed to update balance: %w", err)
	}

	tx.Status = models.StatusCompleted
	if err := s.ProcessTransaction(ctx, tx); err != nil {
		s.logger.Error("failed to update completed transaction", "error", err)
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	details := fmt.Sprintf("Debit transaction %d completed: %s debited from user %d", tx.ID, amount, userID)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, "debit", details); err != nil {
		s.logger.Error("failed to log debit audit", "error", err)
	}

	transactionCounter.WithLabelValues("debit", "success").Inc()
	balanceGauge.WithLabelValues(fmt.Sprintf("%d", userID)).Set(balance.SafeAmount().InexactFloat64())

	return nil
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID, toUserID uint, amount decimal.Decimal, notes string) error {
	timer := prometheus.NewTimer(transactionDuration.WithLabelValues("transfer"))
	defer timer.ObserveDuration()

	if amount.IsNegative() || amount.IsZero() {
		return models.ErrInvalidAmount
	}

	tx := &models.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Type:       models.TypeTransfer,
		Status:     models.StatusPending,
		Notes:      notes,
	}

	if err := tx.Validate(); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Create the transaction first
	if err := s.repo.Create(ctx, tx); err != nil {
		transactionErrors.WithLabelValues("transfer", "creation").Inc()
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// Use the processor to handle the transfer
	if err := s.processor.ProcessTransaction(ctx, tx); err != nil {
		tx.Status = models.StatusFailed
		if updateErr := s.repo.Update(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction status", "error", updateErr)
		}
		transactionErrors.WithLabelValues("transfer", "processing").Inc()
		return fmt.Errorf("failed to process transfer: %w", err)
	}

	// Update transaction status to completed
	tx.Status = models.StatusCompleted
	if err := s.repo.Update(ctx, tx); err != nil {
		s.logger.Error("failed to update completed transaction status", "error", err)
	}

	s.logger.Info("Transfer completed successfully",
		"transaction_id", tx.ID,
		"from_user", fromUserID,
		"to_user", toUserID,
		"amount", amount)

	transactionCounter.WithLabelValues("transfer", "success").Inc()
	transactionAmount.WithLabelValues("transfer").Observe(amount.InexactFloat64())

	return nil
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	tx.Status = models.StatusPending
	if err := s.repo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (s *TransactionService) ProcessTransaction(ctx context.Context, tx *models.Transaction) error {
	if err := s.repo.Update(ctx, tx); err != nil {
		transactionErrors.WithLabelValues(string(tx.Type), "processing_failed").Inc()
		return err
	}
	return nil
}

func (s *TransactionService) GetUserTransactions(ctx context.Context, userID uint) ([]models.Transaction, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *TransactionService) SubmitTransaction(ctx context.Context, tx *models.Transaction) error {
	activeTransactions.WithLabelValues(string(tx.Type)).Inc()
	defer activeTransactions.WithLabelValues(string(tx.Type)).Dec()

	if err := s.CreateTransaction(ctx, tx); err != nil {
		transactionErrors.WithLabelValues(string(tx.Type), "creation").Inc()
		return err
	}

	if err := s.processor.ProcessTransaction(ctx, tx); err != nil {
		tx.Status = models.StatusFailed
		if updateErr := s.ProcessTransaction(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction", "error", updateErr)
			transactionErrors.WithLabelValues(string(tx.Type), "status_update").Inc()
		}
		transactionErrors.WithLabelValues(string(tx.Type), "processing").Inc()
		return fmt.Errorf("failed to process transaction: %w", err)
	}

	transactionCounter.WithLabelValues(string(tx.Type), "success").Inc()
	return nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, transactionID uint) (*models.Transaction, error) {
	tx, err := s.repo.GetByID(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	userID := auth.GetUserIDFromContext(ctx)
	if userID == 0 {
		return nil, fmt.Errorf("unauthorized access")
	}

	if tx.FromUserID != userID && tx.ToUserID != userID {
		return nil, fmt.Errorf("unauthorized access to transaction")
	}

	return tx, nil
}

func (s *TransactionService) Start(ctx context.Context) error {
	return s.processor.Start(ctx)
}

func (s *TransactionService) Stop() {
	s.processor.Stop()
}
