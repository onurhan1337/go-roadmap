package services

import (
	"context"
	"fmt"

	"ledger-link/internal/models"
	"ledger-link/internal/processor"
	"ledger-link/pkg/logger"
)

type TransactionService struct {
	repo      models.TransactionRepository
	processor *processor.TransactionProcessor
	logger    *logger.Logger
}

func NewTransactionService(
	repo models.TransactionRepository,
	balanceSvc models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *TransactionService {
	procConfig := processor.ProcessorConfig{
		WorkerCount:    5,
		QueueSize:      100,
		TransactionSvc: nil,
		BalanceSvc:     balanceSvc,
		AuditSvc:       auditSvc,
		Logger:         logger,
	}

	svc := &TransactionService{
		repo:   repo,
		logger: logger,
	}

	procConfig.TransactionSvc = svc
	svc.processor = processor.NewTransactionProcessor(procConfig)

	return svc
}

func (s *TransactionService) Start(ctx context.Context) error {
	return s.processor.Start(ctx)
}

func (s *TransactionService) Stop() {
	s.processor.Stop()
}

func (s *TransactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	tx.Status = models.StatusPending
	if err := s.repo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (s *TransactionService) ProcessTransaction(ctx context.Context, tx *models.Transaction) error {
	return s.repo.Update(ctx, tx)
}

func (s *TransactionService) GetUserTransactions(ctx context.Context, userID uint) ([]models.Transaction, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *TransactionService) SubmitTransaction(ctx context.Context, tx *models.Transaction) error {
	if err := s.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	if err := s.processor.Submit(tx); err != nil {
		tx.Status = models.StatusFailed
		if updateErr := s.ProcessTransaction(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction", "error", updateErr)
		}
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return nil
}

func (s *TransactionService) GetStatistics() map[string]interface{} {
	return s.processor.GetStatistics()
}