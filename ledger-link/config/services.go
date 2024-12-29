package config

import (
	"context"

	"gorm.io/gorm"

	"ledger-link/internal/repositories"
	"ledger-link/internal/services"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/logger"
)

type ServiceContainer struct {
	db     *gorm.DB
	log    *logger.Logger
	config *Config

	UserRepo        *repositories.UserRepository
	BalanceRepo     *repositories.BalanceRepository
	TransactionRepo *repositories.TransactionRepository
	AuditRepo       *repositories.AuditRepository

	UserService        *services.UserService
	BalanceService     *services.BalanceService
	TransactionService *services.TransactionService
	AuditService       *services.AuditService
	AuthService        *services.AuthService
}

func NewServiceContainer(db *gorm.DB, log *logger.Logger, config *Config) *ServiceContainer {
	container := &ServiceContainer{
		db:     db,
		log:    log,
		config: config,
	}
	container.initRepositories()
	container.initServices()
	return container
}

func (c *ServiceContainer) initRepositories() {
	c.UserRepo = repositories.NewUserRepository(c.db)
	c.BalanceRepo = repositories.NewBalanceRepository(c.db)
	c.TransactionRepo = repositories.NewTransactionRepository(c.db)
	c.AuditRepo = repositories.NewAuditRepository(c.db)
}

func (c *ServiceContainer) initServices() {
	c.AuditService = services.NewAuditService(c.AuditRepo, c.log)
	c.BalanceService = services.NewBalanceService(c.BalanceRepo, c.AuditService, c.log)
	c.UserService = services.NewUserService(c.UserRepo, c.BalanceService, c.AuditService, c.log)
	c.TransactionService = services.NewTransactionService(c.TransactionRepo, c.BalanceService, c.AuditService, c.log)
	tokenMaker := auth.NewJWTMaker(c.config.JWTSecret)
	c.AuthService = services.NewAuthService(c.UserService, tokenMaker, c.log)
}

func (c *ServiceContainer) Start(ctx context.Context) error {
	return nil
}

func (c *ServiceContainer) Stop() {
}