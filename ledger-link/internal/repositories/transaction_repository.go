package repositories

import (
	"context"
	"errors"
	"fmt"

	"ledger-link/internal/models"

	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	result := r.db.WithContext(ctx).Create(tx)
	if result.Error != nil {
		return fmt.Errorf("failed to create transaction: %w", result.Error)
	}
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	var tx models.Transaction
	result := r.db.WithContext(ctx).First(&tx, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", result.Error)
	}
	return &tx, nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Transaction, error) {
	var transactions []models.Transaction
	result := r.db.WithContext(ctx).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&transactions)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", result.Error)
	}
	return transactions, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *models.Transaction) error {
	result := r.db.WithContext(ctx).Save(tx)
	if result.Error != nil {
		return fmt.Errorf("failed to update transaction: %w", result.Error)
	}
	return nil
}