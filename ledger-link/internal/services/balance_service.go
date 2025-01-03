package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"
)

type BalanceService struct {
	repo      models.BalanceRepository
	auditSvc  models.AuditService
	logger    *logger.Logger
	cacheMu   sync.RWMutex
	cache     map[uint]*models.Balance
	locks     sync.Map
}

func NewBalanceService(
	repo models.BalanceRepository,
	auditSvc models.AuditService,
	logger *logger.Logger,
) *BalanceService {
	return &BalanceService{
		repo:     repo,
		auditSvc: auditSvc,
		logger:   logger,
		cache:    make(map[uint]*models.Balance),
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uint) (*models.Balance, error) {
	s.cacheMu.RLock()
	if balance, ok := s.cache[userID]; ok {
		s.cacheMu.RUnlock()
		return balance, nil
	}
	s.cacheMu.RUnlock()

	balance, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if err == models.ErrNotFound {
			balance = &models.Balance{
				UserID:        userID,
				Amount:        0,
				LastUpdatedAt: time.Now(),
			}
			if err := s.repo.Create(ctx, balance); err != nil {
				return nil, fmt.Errorf("failed to create initial balance: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	}

	s.cacheMu.Lock()
	s.cache[userID] = balance
	s.cacheMu.Unlock()

	return balance, nil
}

func (s *BalanceService) UpdateBalance(ctx context.Context, userID uint, amount float64) error {
	lock := s.getLock(userID)
	lock.Lock()
	defer lock.Unlock()

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return err
	}

	oldAmount := balance.SafeAmount()
	newAmount := oldAmount + amount

	if newAmount < 0 {
		return fmt.Errorf("balance cannot be negative")
	}

	balance.UpdateAmount(newAmount)

	if err := s.repo.Update(ctx, balance); err != nil {
		balance.UpdateAmount(oldAmount)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	history := &models.BalanceHistory{
		UserID:    userID,
		OldAmount: oldAmount,
		NewAmount: newAmount,
		CreatedAt: time.Now(),
	}
	if err := s.createBalanceHistory(ctx, history); err != nil {
		s.logger.Error("failed to create balance history", "error", err)
	}

	details := fmt.Sprintf("Balance updated from %.2f to %.2f", oldAmount, newAmount)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeBalance, userID, models.ActionUpdate, details); err != nil {
		s.logger.Error("failed to log balance update", "error", err)
	}

	s.cacheMu.Lock()
	s.cache[userID] = balance
	s.cacheMu.Unlock()

	return nil
}

func (s *BalanceService) LockBalance(ctx context.Context, userID uint) (*sync.Mutex, error) {
	return s.getLock(userID), nil
}

func (s *BalanceService) GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]models.BalanceHistory, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetBalanceHistory(ctx, userID, limit)
}

func (s *BalanceService) createBalanceHistory(ctx context.Context, history *models.BalanceHistory) error {
	return s.repo.CreateBalanceHistory(ctx, history)
}

func (s *BalanceService) getLock(userID uint) *sync.Mutex {
	lock, _ := s.locks.LoadOrStore(userID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

func (s *BalanceService) InvalidateCache(userID uint) {
	s.cacheMu.Lock()
	delete(s.cache, userID)
	s.cacheMu.Unlock()
}