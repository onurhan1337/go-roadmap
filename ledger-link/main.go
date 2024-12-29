package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"ledger-link/config"
	"ledger-link/internal/database"
	"ledger-link/pkg/logger"
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
	db, err := database.InitDB()
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}

	// Initialize service container
	container := config.NewServiceContainer(db, logger, cfg)

	// Create and configure HTTP server
	srv := server.New(cfg, container, logger)

	// Start server in a goroutine
	go func() {
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPIdleTimeout)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
