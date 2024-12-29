package router

import (
	"net/http"
	"time"

	"ledger-link/pkg/logger"
	"ledger-link/pkg/middleware"
)

type Router struct {
	mux    *http.ServeMux
	logger *logger.Logger
}

func NewRouter(logger *logger.Logger) *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger,
	}
}

func (r *Router) Handler() http.Handler {
	handler := middleware.Recovery(r.logger)(r.mux)
	handler = middleware.RequestID()(handler)
	handler = middleware.RateLimit(100, time.Minute)(handler)
	handler = middleware.CORS()(handler)
	handler = middleware.SecurityHeaders()(handler)
	handler = middleware.RequestLogger(r.logger)(handler)

	return handler
}

func (r *Router) Handle(pattern string, handler http.Handler) {
	r.mux.Handle(pattern, handler)
}

func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, handler)
}