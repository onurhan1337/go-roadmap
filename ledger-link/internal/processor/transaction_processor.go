package processor

import (
	"context"
	"fmt"
	"sync"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

type TransactionProcessor struct {
	repo       models.TransactionRepository
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
		repo:       repo,
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

	balance, err := p.balanceSvc.GetBalance(ctx, tx.ToUserID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	newAmount := balance.SafeAmount().Add(tx.Amount)
	if err := p.balanceSvc.UpdateBalance(ctx, tx.ToUserID, newAmount); err != nil {
		return fmt.Errorf("failed to process deposit: %w", err)
	}

	details := fmt.Sprintf("Processed deposit of %s", tx.Amount)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log deposit", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) processWithdrawal(ctx context.Context, tx *models.Transaction) error {
	lock := p.getBalanceLock(tx.FromUserID)
	lock.Lock()
	defer lock.Unlock()

	balance, err := p.balanceSvc.GetBalance(ctx, tx.FromUserID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.SafeAmount().LessThan(tx.Amount) {
		return fmt.Errorf("insufficient funds: available %s, required %s", balance.SafeAmount(), tx.Amount)
	}

	newAmount := balance.SafeAmount().Sub(tx.Amount)
	if err := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, newAmount); err != nil {
		return fmt.Errorf("failed to process withdrawal: %w", err)
	}

	details := fmt.Sprintf("Processed withdrawal of %s", tx.Amount)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log withdrawal", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) processTransfer(ctx context.Context, tx *models.Transaction) error {
	p.logger.Info("Starting transfer process",
		"transaction_id", tx.ID,
		"from_user", tx.FromUserID,
		"to_user", tx.ToUserID,
		"amount", tx.Amount)

	fromLock := p.getBalanceLock(tx.FromUserID)
	toLock := p.getBalanceLock(tx.ToUserID)

	// Acquire locks in a consistent order to prevent deadlocks
	if tx.FromUserID < tx.ToUserID {
		p.logger.Debug("Acquiring locks in order: from -> to")
		fromLock.Lock()
		toLock.Lock()
	} else {
		p.logger.Debug("Acquiring locks in order: to -> from")
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
		p.logger.Debug("Released all locks")
	}()

	// Get sender's balance
	fromBalance, err := p.balanceSvc.GetBalance(ctx, tx.FromUserID)
	if err != nil {
		p.logger.Error("Failed to get sender balance",
			"error", err,
			"user_id", tx.FromUserID)
		return fmt.Errorf("failed to get sender balance: %w", err)
	}
	p.logger.Info("Got sender balance",
		"user_id", tx.FromUserID,
		"current_balance", fromBalance.SafeAmount())

	// Check if sender has sufficient funds
	if fromBalance.SafeAmount().LessThan(tx.Amount) {
		p.logger.Error("Insufficient funds",
			"available", fromBalance.SafeAmount(),
			"required", tx.Amount)
		return fmt.Errorf("insufficient funds: available %s, required %s", fromBalance.SafeAmount(), tx.Amount)
	}

	// Get receiver's balance
	toBalance, err := p.balanceSvc.GetBalance(ctx, tx.ToUserID)
	if err != nil {
		p.logger.Error("Failed to get receiver balance",
			"error", err,
			"user_id", tx.ToUserID)
		return fmt.Errorf("failed to get receiver balance: %w", err)
	}
	p.logger.Info("Got receiver balance",
		"user_id", tx.ToUserID,
		"current_balance", toBalance.SafeAmount())

	// Calculate new balances
	newFromAmount := fromBalance.SafeAmount().Sub(tx.Amount)
	newToAmount := toBalance.SafeAmount().Add(tx.Amount)

	p.logger.Info("Calculated new balances",
		"sender_old_balance", fromBalance.SafeAmount(),
		"sender_new_balance", newFromAmount,
		"receiver_old_balance", toBalance.SafeAmount(),
		"receiver_new_balance", newToAmount)

	// Update sender's balance in DB
	if err := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, newFromAmount); err != nil {
		p.logger.Error("Failed to update sender balance",
			"error", err,
			"user_id", tx.FromUserID,
			"new_amount", newFromAmount)
		return fmt.Errorf("failed to update sender balance: %w", err)
	}
	p.logger.Info("Updated sender balance successfully")

	// Update receiver's balance in DB
	if err := p.balanceSvc.UpdateBalance(ctx, tx.ToUserID, newToAmount); err != nil {
		// Rollback sender's balance
		p.logger.Error("Failed to update receiver balance, rolling back sender's balance",
			"error", err,
			"user_id", tx.ToUserID,
			"new_amount", newToAmount)
		if rbErr := p.balanceSvc.UpdateBalance(ctx, tx.FromUserID, fromBalance.SafeAmount()); rbErr != nil {
			p.logger.Error("Failed to rollback sender balance",
				"error", rbErr,
				"user_id", tx.FromUserID)
		}
		return fmt.Errorf("failed to update receiver balance: %w", err)
	}
	p.logger.Info("Updated receiver balance successfully")

	details := fmt.Sprintf("Processed transfer of %s from %d to %d", tx.Amount, tx.FromUserID, tx.ToUserID)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("Failed to log transfer audit", "error", err)
	}

	p.logger.Info("Transfer completed successfully",
		"transaction_id", tx.ID,
		"from_user", tx.FromUserID,
		"to_user", tx.ToUserID,
		"amount", tx.Amount)

	return nil
}

func (p *TransactionProcessor) getBalanceLock(userID uint) *sync.Mutex {
	lock, _ := p.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (p *TransactionProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting transaction processor")
	return nil
}

func (p *TransactionProcessor) Stop() {
	p.logger.Info("stopping transaction processor")
}
