package config

import (
	"fmt"
	"ledger-link/internal/handlers"
	"ledger-link/internal/repositories"
	"ledger-link/internal/services"
	"ledger-link/pkg/auth"
	"ledger-link/pkg/cache"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/redis"

	"gorm.io/gorm"
)

type ServiceContainer struct {
	// Services
	AuthService        *services.AuthService
	UserService        *services.UserService
	TransactionService *services.TransactionService
	BalanceService     *services.BalanceService
	AuditService       *services.AuditService

	// Handlers
	AuthHandler        *handlers.AuthHandler
	UserHandler        *handlers.UserHandler
	TransactionHandler *handlers.TransactionHandler
	BalanceHandler     *handlers.BalanceHandler

	// Redis
	CacheService *cache.CacheService
}

func NewServiceContainer(db *gorm.DB, logger *logger.Logger, cfg *Config) (*ServiceContainer, error) {
	// Initialize Redis client
	redisClient, err := redis.NewRedisClient(&redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}

	// Initialize cache service
	cacheService := cache.NewCacheService(redisClient)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	balanceRepo := repositories.NewBalanceRepository(db)
	auditRepo := repositories.NewAuditLogRepository(db)

	// Initialize JWT token maker
	tokenMaker := auth.NewJWTMaker(cfg.JWT.SecretKey)

	// Initialize services
	auditSvc := services.NewAuditService(auditRepo, logger)
	balanceSvc := services.NewBalanceService(balanceRepo, auditSvc, logger, cacheService)
	userSvc := services.NewUserService(userRepo, balanceSvc, auditSvc, logger)
	authSvc := services.NewAuthService(userSvc, tokenMaker, logger)
	transactionSvc := services.NewTransactionService(transactionRepo, balanceSvc, auditSvc, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authSvc, logger)
	userHandler := handlers.NewUserHandler(userSvc, logger)
	transactionHandler := handlers.NewTransactionHandler(transactionSvc, logger)
	balanceHandler := handlers.NewBalanceHandler(balanceSvc, logger, nil) // Using default config

	return &ServiceContainer{
		// Services
		AuthService:        authSvc,
		UserService:        userSvc,
		TransactionService: transactionSvc,
		BalanceService:     balanceSvc,
		AuditService:       auditSvc,

		// Handlers
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		TransactionHandler: transactionHandler,
		BalanceHandler:     balanceHandler,

		// Redis
		CacheService: cacheService,
	}, nil
}
