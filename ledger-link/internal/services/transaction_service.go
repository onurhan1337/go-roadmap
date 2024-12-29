package services

import (
	"context"
	"fmt"

	"ledger-link/internal/models"
	"ledger-link/internal/processor"
	"ledger-link/pkg/logger"
)

type TransactionService struct {
	repo        models.TransactionRepository
	processor   *processor.TransactionProcessor
	balanceSvc  models.BalanceService
	auditSvc    models.AuditService
	logger      *logger.Logger
}

func NewTransactionService(
	repo models.TransactionRepository,
	balanceSvc models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *TransactionService {
	svc := &TransactionService{
		repo:       repo,
		balanceSvc: balanceSvc,
		auditSvc:   auditSvc,
		logger:     logger,
	}

	procConfig := processor.ProcessorConfig{
		WorkerCount:    5,
		QueueSize:      100,
		TransactionSvc: svc,
		BalanceSvc:     balanceSvc,
		AuditSvc:       auditSvc,
		Logger:         logger,
	}

	svc.processor = processor.NewTransactionProcessor(procConfig)
	return svc
}

func (s *TransactionService) Start(ctx context.Context) error {
	return s.processor.Start(ctx)
}

func (s *TransactionService) Stop() {
	s.processor.Stop()
}

func (s *TransactionService) Credit(ctx context.Context, userID uint, amount float64, notes string) error {
	if amount <= 0 {
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

	balance, err := s.balanceSvc.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if err := s.CreateTransaction(ctx, tx); err != nil {
		return err
	}

	if err := balance.AddAmount(amount); err != nil {
		tx.Status = models.StatusFailed
		if updateErr := s.ProcessTransaction(ctx, tx); updateErr != nil {
			s.logger.Error("failed to update failed transaction", "error", updateErr)
		}
		return fmt.Errorf("failed to credit amount: %w", err)
	}

	if err := s.balanceSvc.UpdateBalance(ctx, userID, balance.SafeAmount()); err != nil {
		if rbErr := balance.SubtractAmount(amount); rbErr != nil {
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

	details := fmt.Sprintf("Credit transaction %d completed: %f credited to user %d", tx.ID, amount, userID)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, "credit", details); err != nil {
		s.logger.Error("failed to log credit audit", "error", err)
	}

	return nil
}

func (s *TransactionService) Debit(ctx context.Context, userID uint, amount float64, notes string) error {
	if amount <= 0 {
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

	if balance.SafeAmount() < amount {
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

	details := fmt.Sprintf("Debit transaction %d completed: %f debited from user %d", tx.ID, amount, userID)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, "debit", details); err != nil {
		s.logger.Error("failed to log debit audit", "error", err)
	}

	return nil
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID, toUserID uint, amount float64, notes string) error {
	if amount <= 0 {
		return models.ErrInvalidAmount
	}

	if fromUserID == toUserID {
		return fmt.Errorf("cannot transfer to the same account")
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

	return s.SubmitTransaction(ctx, tx)
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