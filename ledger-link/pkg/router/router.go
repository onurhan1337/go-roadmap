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

func New() *Router {
	return &Router{
		mux:    http.NewServeMux(),
		logger: logger.New("info"),
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

func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.HandleMethod("GET", path, handler)
}

func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.HandleMethod("POST", path, handler)
}

func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.HandleMethod("PUT", path, handler)
}

func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.HandleMethod("DELETE", path, handler)
}

func (r *Router) HandleMethod(method, path string, handler http.HandlerFunc) {
	r.mux.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, req)
	})
}