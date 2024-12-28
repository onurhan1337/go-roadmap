package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"ledger-link/config"
	"ledger-link/internal/database"
	"ledger-link/pkg/logger"

	"github.com/joho/godotenv"
)

type App struct {
	log *logger.Logger
	cfg *config.Config
}

func NewApp(cfg *config.Config) *App {
	return &App{
		log: logger.New(cfg.LogLevel),
		cfg: cfg,
	}
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			a.log.Info("received signal", "signal", sig)
			cancel()
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		a.log.Error("error during shutdown", "error", err)
		return err
	}

	a.log.Info("shutdown complete")
	return nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get underlying *sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	app := NewApp(cfg)
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
