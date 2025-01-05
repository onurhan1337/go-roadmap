package server

import (
	"context"
	"fmt"
	"net/http"

	"ledger-link/config"
	"ledger-link/pkg/logger"
	"ledger-link/pkg/middleware"
	"ledger-link/pkg/ratelimit"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
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

func (s *Server) setupRoutes() {
	// Add metrics endpoint
	s.server.Handler = promhttp.Handler()
}

func (s *Server) WithRateLimiter(redisClient *redis.Client) *Server {
	rateLimiter := ratelimit.NewRateLimiter(redisClient)
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for metrics endpoint
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			// Apply rate limiting based on path
			var handler http.Handler
			switch r.URL.Path {
			case "/api/v1/auth/login":
				handler = rateLimiter.Limit("login", middleware.LoginRateLimit)(next)
			case "/api/v1/auth/register":
				handler = rateLimiter.Limit("register", middleware.RegisterRateLimit)(next)
			case "/api/v1/transactions":
				handler = rateLimiter.Limit("transaction", middleware.TransactionRateLimit)(next)
			case "/api/v1/balances/transfer":
				handler = rateLimiter.Limit("balance", middleware.BalanceOperationRateLimit)(next)
			default:
				handler = next
			}
			handler.ServeHTTP(w, r)
		})
	})
	return s
}
