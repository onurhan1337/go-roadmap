package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ledger-link/internal/models"
)

type BalanceRepository struct {
	db *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) *BalanceRepository {
	return &BalanceRepository{
		db: db,
	}
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID uint) (*models.Balance, error) {
	var balance models.Balance
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	return &balance, nil
}

func (r *BalanceRepository) Create(ctx context.Context, balance *models.Balance) error {
	if err := r.db.WithContext(ctx).Create(balance).Error; err != nil {
		return fmt.Errorf("failed to create balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) Update(ctx context.Context, balance *models.Balance) error {
	if err := r.db.WithContext(ctx).Save(balance).Error; err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	return nil
}

func (r *BalanceRepository) GetBalanceHistory(ctx context.Context, userID uint, limit int) ([]models.BalanceHistory, error) {
	var history []models.BalanceHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

func (r *BalanceRepository) GetBalanceHistoryAfterTime(ctx context.Context, userID uint, timestamp time.Time) ([]models.BalanceHistory, error) {
	var history []models.BalanceHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ?", userID, timestamp).
		Order("created_at ASC").
		Find(&history).Error
	return history, err
}

func (r *BalanceRepository) CreateBalanceHistory(ctx context.Context, history *models.BalanceHistory) error {
	if err := r.db.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to create balance history: %w", err)
	}
	return nil
}
