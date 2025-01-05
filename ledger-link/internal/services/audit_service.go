package services

import (
	"context"

	"ledger-link/internal/models"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type AuditService struct {
	repo   models.AuditLogRepository
	logger *logger.Logger
}

func NewAuditService(repo models.AuditLogRepository, logger *logger.Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

func (s *AuditService) LogAction(ctx context.Context, entityType string, entityID uint, action string, details string) error {
	userID := auth.GetUserIDFromContext(ctx)
	if userID == 0 {
		s.logger.Warn("no user ID found in context for audit log", "entityType", entityType, "entityID", entityID)
		userID = entityID
	}

	log := &models.AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Details:    details,
		UserID:     userID,
	}
	return s.repo.Create(ctx, log)
}

func (s *AuditService) GetEntityAuditLog(ctx context.Context, entityType string, entityID uint) ([]models.AuditLog, error) {
	return s.repo.GetByEntityID(ctx, entityType, entityID)
}
