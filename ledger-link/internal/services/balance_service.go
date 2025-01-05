package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ledger-link/internal/models"
	"ledger-link/pkg/cache"
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
	cache    *cache.CacheService
	locks    sync.Map
}

func NewBalanceService(
	repo models.BalanceRepository,
	auditSvc models.AuditService,
	logger *logger.Logger,
	cache *cache.CacheService,
) *BalanceService {
	return &BalanceService{
		repo:     repo,
		auditSvc: auditSvc,
		logger:   logger,
		cache:    cache,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID uint) (*models.Balance, error) {
	timer := prometheus.NewTimer(balanceUpdateDuration.WithLabelValues("get"))
	defer timer.ObserveDuration()

	cacheKey := cache.BuildKey(cache.KeyBalance, userID)
	var balance *models.Balance

	if err := s.cache.Get(ctx, cacheKey, &balance); err == nil {
		balanceOperations.WithLabelValues("get", "cache_hit").Inc()
		balanceDistribution.WithLabelValues("current").Observe(balance.SafeAmount())
		return balance, nil
	}

	balance, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if err == models.ErrNotFound {
			balance = &models.Balance{
				UserID:        userID,
				Amount:        0,
				LastUpdatedAt: time.Now(),
			}
			if err := s.CreateInitialBalance(ctx, balance); err != nil {
				balanceOperations.WithLabelValues("get", "failure").Inc()
				return nil, fmt.Errorf("failed to create initial balance: %w", err)
			}
		} else {
			balanceOperations.WithLabelValues("get", "failure").Inc()
			return nil, fmt.Errorf("failed to get balance: %w", err)
		}
	}

	if err := s.cache.Set(ctx, cacheKey, balance, cache.ShortTerm); err != nil {
		s.logger.Error("failed to cache balance", "error", err)
	}

	balanceOperations.WithLabelValues("get", "db_hit").Inc()
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

	cacheKey := fmt.Sprintf("balance:%d", userID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		s.logger.Error("failed to invalidate balance cache", "error", err)
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

func (s *BalanceService) CreateInitialBalance(ctx context.Context, balance *models.Balance) error {
	lock := s.getLock(balance.UserID)
	lock.Lock()
	defer lock.Unlock()

	_, err := s.repo.GetByUserID(ctx, balance.UserID)
	if err == nil {
		return nil
	} else if err != models.ErrNotFound {
		return fmt.Errorf("failed to check existing balance: %w", err)
	}

	if err := s.repo.Create(ctx, balance); err != nil {
		return fmt.Errorf("failed to create initial balance: %w", err)
	}

	cacheKey := cache.BuildKey(cache.KeyBalance, balance.UserID)
	if err := s.cache.Set(ctx, cacheKey, balance, cache.MediumTerm); err != nil {
		s.logger.Error("failed to cache initial balance", "error", err)
	}

	details := fmt.Sprintf("Initial balance created with amount %.2f", balance.Amount)
	if err := s.auditSvc.LogAction(ctx, models.EntityTypeBalance, balance.UserID, models.ActionCreate, details); err != nil {
		s.logger.Error("failed to log initial balance creation", "error", err)
	}

	return nil
}
