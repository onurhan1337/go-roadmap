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

	router := router.New()

	// Auth routes
	authHandler := handlers.NewAuthHandler(container.AuthService, logger)
	router.POST("/api/v1/auth/register", authHandler.Register)
	router.POST("/api/v1/auth/login", authHandler.Login)
	router.POST("/api/v1/auth/refresh", authHandler.RefreshToken)

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