package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	balanceOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledger_balance_operations_total",
			Help: "Total number of balance operations",
		},
		[]string{"operation", "status"},
	)

	balanceUpdateDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_balance_update_duration_seconds",
			Help:    "Balance update operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	balanceDistribution = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledger_balance_distribution",
			Help:    "Distribution of user balances",
			Buckets: []float64{0, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"type"},
	)
)

type BalanceService struct {
	repo     models.BalanceRepository
	auditSvc models.AuditService
	logger   *logger.Logger
	cacheMu  sync.RWMutex
	cache    map[uint]*models.Balance
	locks    sync.Map
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
	timer := prometheus.NewTimer(balanceUpdateDuration.WithLabelValues("get"))
	defer timer.ObserveDuration()

	s.cacheMu.RLock()
	if balance, ok := s.cache[userID]; ok {
		s.cacheMu.RUnlock()
		balanceOperations.WithLabelValues("get", "success").Inc()
		balanceDistribution.WithLabelValues("current").Observe(balance.SafeAmount())
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
				balanceOperations.WithLabelValues("get", "failure").Inc()
				return nil, fmt.Errorf("failed to create initial balance: %w", err)
			}
		} else {
			balanceOperations.WithLabelValues("get", "failure").Inc()
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	}

	s.cacheMu.Lock()
	s.cache[userID] = balance
	s.cacheMu.Unlock()

	balanceOperations.WithLabelValues("get", "success").Inc()
	balanceDistribution.WithLabelValues("current").Observe(balance.SafeAmount())
	return balance, nil
}

func (s *BalanceService) UpdateBalance(ctx context.Context, userID uint, amount float64) error {
	timer := prometheus.NewTimer(balanceUpdateDuration.WithLabelValues("update"))
	defer timer.ObserveDuration()

	lock := s.getLock(userID)
	lock.Lock()
	defer lock.Unlock()

	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		balanceOperations.WithLabelValues("update", "failure").Inc()
		return err
	}

	oldAmount := balance.SafeAmount()
	newAmount := amount

	if newAmount < 0 {
		balanceOperations.WithLabelValues("update", "failure").Inc()
		return fmt.Errorf("balance cannot be negative")
	}

	balance.UpdateAmount(newAmount)

	if err := s.repo.Update(ctx, balance); err != nil {
		balance.UpdateAmount(oldAmount)
		balanceOperations.WithLabelValues("update", "failure").Inc()
		return fmt.Errorf("failed to update balance: %w", err)
	}

	balanceOperations.WithLabelValues("update", "success").Inc()
	balanceDistribution.WithLabelValues("current").Observe(newAmount)

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
	history, err := s.repo.GetBalanceHistory(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}

	return history, nil
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

func (s *BalanceService) GetBalanceAtTime(ctx context.Context, userID uint, timestamp time.Time) (*models.Balance, error) {
	currentBalance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current balance: %w", err)
	}

	balance := &models.Balance{
		UserID:        currentBalance.UserID,
		Amount:        currentBalance.SafeAmount(),
		LastUpdatedAt: currentBalance.LastUpdatedAt,
	}

	history, err := s.repo.GetBalanceHistory(ctx, userID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", err)
	}

	for i := len(history) - 1; i >= 0; i-- {
		if history[i].CreatedAt.After(timestamp) {
			balance.Amount = history[i].OldAmount
			balance.LastUpdatedAt = history[i].CreatedAt
		} else {
			break
		}
	}

	return balance, nil
}
