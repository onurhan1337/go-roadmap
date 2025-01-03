package server

import (
	"context"
	"fmt"
	"net/http"

	"ledger-link/config"
	"ledger-link/pkg/logger"
)

type Server struct {
	server *http.Server
	logger *logger.Logger
}

func New(cfg config.ServerConfig, handler http.Handler, logger *logger.Logger) *Server {
	addr := fmt.Sprintf("%s:%s", cfg.Address, cfg.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return &Server{
		server: server,
		logger: logger,
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Use(middleware func(http.Handler) http.Handler) {
	s.server.Handler = middleware(s.server.Handler)
}
