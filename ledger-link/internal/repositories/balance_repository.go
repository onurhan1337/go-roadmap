package repositories

import (
	"context"
	"errors"
	"fmt"

	"ledger-link/internal/models"

	"gorm.io/gorm"
)

type BalanceRepository struct {
	db *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) Create(ctx context.Context, balance *models.Balance) error {
	result := r.db.WithContext(ctx).Create(balance)
	if result.Error != nil {
		return fmt.Errorf("failed to create balance: %w", result.Error)
	}
	return nil
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID uint) (*models.Balance, error) {
	var balance models.Balance
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get balance: %w", result.Error)
	}
	return &balance, nil
}

func (r *BalanceRepository) Update(ctx context.Context, balance *models.Balance) error {
	result := r.db.WithContext(ctx).Save(balance)
	if result.Error != nil {
		return fmt.Errorf("failed to update balance: %w", result.Error)
	}
	return nil
}

func (r *BalanceRepository) GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]models.BalanceHistory, error) {
	var history []models.BalanceHistory
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get balance history: %w", result.Error)
	}
	return history, nil
}

func (r *BalanceRepository) CreateBalanceHistory(ctx context.Context, history *models.BalanceHistory) error {
	result := r.db.WithContext(ctx).Create(history)
	if result.Error != nil {
		return fmt.Errorf("failed to create balance history: %w", result.Error)
	}
	return nil
}