package server

import (
	"context"
	"fmt"
	"net/http"

	"ledger-link/config"
	"ledger-link/internal/handlers"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/router"
)

type Server struct {
	cfg       *config.Config
	container *config.ServiceContainer
	logger    *logger.Logger
	server    *http.Server
}

func New(cfg *config.Config, container *config.ServiceContainer, logger *logger.Logger) *Server {
	s := &Server{
		cfg:       cfg,
		container: container,
		logger:    logger,
	}

	router := router.New(container.AuthService)

	// Create handler configs
	balanceConfig := handlers.DefaultBalanceHandlerConfig()

	// Initialize handlers with configs
	balanceHandler := handlers.NewBalanceHandler(container.BalanceService, logger, balanceConfig)
	authHandler := handlers.NewAuthHandler(container.AuthService, logger)
	transactionHandler := handlers.NewTransactionHandler(container.TransactionService, logger)
	userHandler := handlers.NewUserHandler(container.UserService, logger)

	// Auth routes
	router.POST("/api/v1/auth/register", authHandler.Register)
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

	// User routes
	router.GET("/api/v1/users", userHandler.GetUsers)
	router.GET("/api/v1/users/{id}", userHandler.GetUser)
	router.PUT("/api/v1/users/{id}", userHandler.UpdateUser)
	router.DELETE("/api/v1/users/{id}", userHandler.DeleteUser)

	// Transaction routes
	router.POST("/api/v1/transactions/credit", transactionHandler.HandleCredit)
	router.POST("/api/v1/transactions/debit", transactionHandler.HandleDebit)
	router.POST("/api/v1/transactions/transfer", transactionHandler.HandleTransfer)
	router.GET("/api/v1/transactions/history", transactionHandler.HandleGetTransactionHistory)
	router.GET("/api/v1/transactions/{id}", transactionHandler.HandleGetTransaction)

	// Balance routes
	router.GET("/api/v1/balances/current", balanceHandler.GetCurrentBalance)
	router.GET("/api/v1/balances/historical", balanceHandler.GetHistoricalBalances)
	router.GET("/api/v1/balances/at-time", balanceHandler.GetBalanceAtTime)

	s.server = &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router.Handler(),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	return s
}

func (s *Server) Start() error {
	s.logger.Info("starting server", "port", s.cfg.HTTPPort)
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
