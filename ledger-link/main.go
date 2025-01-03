package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ledger-link/config"
	"ledger-link/internal/database"
	"ledger-link/internal/router"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/middleware"
	"ledger-link/pkg/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Connect to database
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}

	// Initialize service container
	container := config.NewServiceContainer(db, logger, cfg)

	// Setup middleware
	authMiddleware := middleware.NewAuthMiddleware(container.AuthService, logger)
	rbacMiddleware := middleware.NewRBACMiddleware(logger)
	errorMiddleware := middleware.NewErrorMiddleware(logger)
	metricsMiddleware := middleware.NewMetricsMiddleware(logger)

	// Initialize router with RBAC
	r := router.NewRouter(
		container.AuthHandler,
		container.UserHandler,
		container.TransactionHandler,
		container.BalanceHandler,
		authMiddleware,
		rbacMiddleware,
	)

	// Create and configure HTTP server with the router
	srv := server.New(cfg.Server, r, logger)

	// Apply global middleware
	srv.Use(middleware.Chain(
		errorMiddleware.HandleError,
		metricsMiddleware.TrackPerformance,
	))

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", "address", cfg.Server.Address, "port", cfg.Server.Port)
		if err := srv.Start(); err != nil {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.HTTPIdleTimeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
