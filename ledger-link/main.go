package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledger-link/config"
	"ledger-link/internal/database"
	"ledger-link/internal/router"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/middleware"
	"ledger-link/pkg/ratelimit"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	// Initialize database
	db, err := database.InitDB(cfg.Database)
	if err != nil {
		log.Fatal("failed to initialize database", "error", err)
	}

	// Initialize service container
	container, err := config.NewServiceContainer(db, log, cfg)
	if err != nil {
		log.Fatal("failed to initialize service container", "error", err)
	}

	// Initialize router with handlers and middleware
	router := router.NewRouter(
		container.AuthHandler,
		container.UserHandler,
		container.TransactionHandler,
		container.BalanceHandler,
		middleware.NewAuthMiddleware(container.AuthService, log),
		middleware.NewRBACMiddleware(log),
		ratelimit.NewRateLimiter(container.CacheService.RedisClient),
	)

	// Wrap router with middleware chain
	handler := middleware.Chain(
		router,
		middleware.MetricsMiddleware,
		middleware.LoggingMiddleware(log),
		middleware.RecoveryMiddleware(log),
	)

	// Create mux for metrics endpoint
	mux := http.NewServeMux()

	// Add metrics endpoint without authentication
	mux.Handle("/metrics", promhttp.Handler())

	// Add health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Add main application handler
	mux.Handle("/", handler)

	// Create server with proper handler
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		IdleTimeout:  cfg.Server.HTTPIdleTimeout,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	// Start server
	go func() {
		log.Info("starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown", "error", err)
	}

	log.Info("server exited properly")
}
