package repositories

import (
	"context"
	"fmt"

	"ledger-link/internal/models"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	result := r.db.WithContext(ctx).Create(log)
	if result.Error != nil {
		return fmt.Errorf("failed to create audit log: %w", result.Error)
	}
	return nil
}

func (r *AuditRepository) GetByEntityID(ctx context.Context, entityType string, entityID uint) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.db.WithContext(ctx).Where("entity_type = ? AND entity_id = ?", entityType, entityID).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	return logs, nil
}