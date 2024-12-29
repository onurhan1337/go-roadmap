package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ledger-link/internal/models"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.WithContext(ctx).First(&transaction, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uint) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.WithContext(ctx).Where("from_user_id = ? OR to_user_id = ?", userID, userID).Order("created_at desc").Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user transactions: %w", err)
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