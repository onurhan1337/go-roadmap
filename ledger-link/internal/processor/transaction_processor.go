package processor

import (
	"context"
	"fmt"
	"sync"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

type TransactionProcessor struct {
	workerCount    int
	jobQueue       chan *models.Transaction
	transactionSvc models.TransactionService
	balanceSvc     models.BalanceService
	auditSvc       models.AuditService
	logger         *logger.Logger
	wg             sync.WaitGroup
	done           chan struct{}
	balanceLocks   sync.Map
	stats          *models.TransactionStats
}

type ProcessorConfig struct {
	WorkerCount     int
	QueueSize       int
	TransactionSvc  models.TransactionService
	BalanceSvc      models.BalanceService
	AuditSvc        models.AuditService
	Logger          *logger.Logger
}

func NewTransactionProcessor(cfg ProcessorConfig) *TransactionProcessor {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 5
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}

	return &TransactionProcessor{
		workerCount:    cfg.WorkerCount,
		jobQueue:       make(chan *models.Transaction, cfg.QueueSize),
		transactionSvc: cfg.TransactionSvc,
		balanceSvc:     cfg.BalanceSvc,
		auditSvc:       cfg.AuditSvc,
		logger:         cfg.Logger,
		done:           make(chan struct{}),
		stats:          models.NewTransactionStats(),
	}
}

func (p *TransactionProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting transaction processor", "workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	go func() {
		<-ctx.Done()
		p.logger.Info("shutting down transaction processor")
		close(p.jobQueue)
		close(p.done)
	}()

	return nil
}

func (p *TransactionProcessor) Stop() {
	p.wg.Wait()
	p.logger.Info("transaction processor stopped")
}

func (p *TransactionProcessor) Submit(tx *models.Transaction) error {
	select {
	case p.jobQueue <- tx:
		return nil
	case <-p.done:
		return fmt.Errorf("transaction processor is shutting down")
	default:
		return fmt.Errorf("transaction queue is full")
	}
}

func (p *TransactionProcessor) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	p.logger.Info("starting worker", "worker_id", id)

	for {
		select {
		case tx, ok := <-p.jobQueue:
			if !ok {
				p.logger.Info("worker shutting down", "worker_id", id)
				return
			}
			if err := p.processTransaction(ctx, tx); err != nil {
				p.logger.Error("failed to process transaction",
					"error", err,
					"transaction_id", tx.ID,
					"worker_id", id,
				)
			}
		case <-ctx.Done():
			p.logger.Info("worker context cancelled", "worker_id", id)
			return
		}
	}
}

func (p *TransactionProcessor) getBalanceLock(userID uint) *sync.Mutex {
	lock, _ := p.balanceLocks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (p *TransactionProcessor) processTransaction(ctx context.Context, tx *models.Transaction) error {
	p.stats.IncrementTotal()

	var firstLock, secondLock *sync.Mutex
	if tx.FromUserID < tx.ToUserID {
		firstLock = p.getBalanceLock(tx.FromUserID)
		secondLock = p.getBalanceLock(tx.ToUserID)
	} else {
		firstLock = p.getBalanceLock(tx.ToUserID)
		secondLock = p.getBalanceLock(tx.FromUserID)
	}

	firstLock.Lock()
	secondLock.Lock()
	defer firstLock.Unlock()
	defer secondLock.Unlock()

	fromBalance, err := p.balanceSvc.GetBalance(ctx, tx.FromUserID)
	if err != nil {
		p.stats.IncrementFailed()
		return fmt.Errorf("failed to get from user balance: %w", err)
	}

	if fromBalance.SafeAmount() < tx.Amount {
			tx.Status = models.StatusFailed
			p.stats.IncrementFailed()
			if err := p.transactionSvc.ProcessTransaction(ctx, tx); err != nil {
				return fmt.Errorf("failed to update failed transaction: %w", err)
			}
			return fmt.Errorf("insufficient funds")
	}

	if err := fromBalance.SubtractAmount(tx.Amount); err != nil {
		p.stats.IncrementFailed()
		return fmt.Errorf("failed to update from user balance: %w", err)
	}

	toBalance, err := p.balanceSvc.GetBalance(ctx, tx.ToUserID)
	if err != nil {
		if rbErr := fromBalance.AddAmount(tx.Amount); rbErr != nil {
			p.logger.Error("failed to rollback balance update", "error", rbErr)
		}
		p.stats.IncrementFailed()
		return fmt.Errorf("failed to get to user balance: %w", err)
	}

	if err := toBalance.AddAmount(tx.Amount); err != nil {
		if rbErr := fromBalance.AddAmount(tx.Amount); rbErr != nil {
			p.logger.Error("failed to rollback balance update", "error", rbErr)
		}
		p.stats.IncrementFailed()
		return fmt.Errorf("failed to update to user balance: %w", err)
	}

	tx.Status = models.StatusCompleted
	if err := p.transactionSvc.ProcessTransaction(ctx, tx); err != nil {
		p.stats.IncrementFailed()
		return fmt.Errorf("failed to update completed transaction: %w", err)
	}

	p.stats.IncrementSuccessful()
	p.stats.AddAmount(tx.Amount)

	details := fmt.Sprintf("Transaction %d completed: %f transferred from user %d to user %d",
		tx.ID, tx.Amount, tx.FromUserID, tx.ToUserID)
	if err := p.auditSvc.LogAction(ctx, models.EntityTypeTransaction, tx.ID, models.ActionUpdate, details); err != nil {
		p.logger.Error("failed to log transaction audit", "error", err)
	}

	return nil
}

func (p *TransactionProcessor) GetStatistics() map[string]interface{} {
	return p.stats.GetStats()
}