package processor

import (
	"context"
	"fmt"
	"sync"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

type TransactionProcessor struct {
	repo      models.TransactionRepository
	balanceSvc models.BalanceService
	auditSvc   models.AuditService
	logger     *logger.Logger
	locks      sync.Map
}

func NewTransactionProcessor(
	repo models.TransactionRepository,
	balanceSvc models.BalanceService,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *TransactionProcessor {
	return &TransactionProcessor{
		repo:      repo,
		balanceSvc: balanceSvc,
		auditSvc:   auditSvc,
		logger:     logger,
	}
}

func (p *TransactionProcessor) ProcessTransaction(ctx context.Context, tx *models.Transaction) error {
	var err error

	switch tx.Type {
	case models.TypeDeposit:
		err = p.processDeposit(ctx, tx)
	case models.TypeWithdrawal:
		err = p.processWithdrawal(ctx, tx)
	case models.TypeTransfer:
		err = p.processTransfer(ctx, tx)
	default:
		return fmt.Errorf("unsupported transaction type: %s", tx.Type)
	}

	if err != nil {
		tx.Status = models.StatusFailed
		if updateErr := p.repo.Update(ctx, tx); updateErr != nil {
			p.logger.Error("failed to update transaction status", "error", updateErr)
		}
		return err
	}

	tx.Status = models.StatusCompleted
	if err := p.repo.Update(ctx, tx); err != nil {
		p.logger.Error("failed to update transaction status", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) processDeposit(ctx context.Context, tx *models.Transaction) error {
	lock := p.getBalanceLock(tx.ToUserID)
	lock.Lock()
	defer lock.Unlock()

	if err := p.balanceSvc.UpdateBalance(ctx, tx.ToUserID, tx.Amount); err != nil {
		return fmt.Errorf("failed to process deposit: %w", err)
	}

	details := fmt.Sprintf("Processed deposit of %.2f", tx.Amount)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log deposit", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) processWithdrawal(ctx context.Context, tx *models.Transaction) error {
	lock := p.getBalanceLock(tx.FromUserID)
	lock.Lock()
	defer lock.Unlock()

	if err := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, -tx.Amount); err != nil {
		return fmt.Errorf("failed to process withdrawal: %w", err)
	}

	details := fmt.Sprintf("Processed withdrawal of %.2f", tx.Amount)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log withdrawal", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) processTransfer(ctx context.Context, tx *models.Transaction) error {
	fromLock := p.getBalanceLock(tx.FromUserID)
	toLock := p.getBalanceLock(tx.ToUserID)

	// Lock in a consistent order to prevent deadlocks
	if tx.FromUserID < tx.ToUserID {
		fromLock.Lock()
		toLock.Lock()
	} else {
		toLock.Lock()
		fromLock.Lock()
	}
	defer func() {
		if tx.FromUserID < tx.ToUserID {
			toLock.Unlock()
			fromLock.Unlock()
		} else {
			fromLock.Unlock()
			toLock.Unlock()
		}
	}()

	if err := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, -tx.Amount); err != nil {
		return fmt.Errorf("failed to debit sender: %w", err)
	}

	if err := p.balanceSvc.UpdateBalance(ctx, tx.ToUserID, tx.Amount); err != nil {
		// Rollback the debit if crediting fails
		if rollbackErr := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, tx.Amount); rollbackErr != nil {
			p.logger.Error("failed to rollback debit", "error", rollbackErr)
		}
		return fmt.Errorf("failed to credit receiver: %w", err)
	}

	details := fmt.Sprintf("Processed transfer of %.2f from %d to %d", tx.Amount, tx.FromUserID, tx.ToUserID)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log transfer", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) getBalanceLock(userID uint) *sync.Mutex {
	lock, _ := p.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}