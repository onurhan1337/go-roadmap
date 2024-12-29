package router

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ledger-link/pkg/logger"
	"ledger-link/pkg/middleware"
)

type contextKey string

const (
	PathParamsKey contextKey = "path_params"
)

type Router struct {
	routes []*route
	logger *logger.Logger
}

type route struct {
	method  string
	regex   *regexp.Regexp
	handler http.HandlerFunc
	params  []string
}

func New() *Router {
	return &Router{
		routes: make([]*route, 0),
		logger: logger.New("info"),
	}
}

func (r *Router) Handler() http.Handler {
	var handler http.Handler = http.HandlerFunc(r.ServeHTTP)
	handler = middleware.Recovery(r.logger)(handler)
	handler = middleware.RequestID()(handler)
	handler = middleware.RateLimit(100, time.Minute)(handler)
	handler = middleware.CORS()(handler)
	handler = middleware.SecurityHeaders()(handler)
	handler = middleware.RequestLogger(r.logger)(handler)

	return handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		if route.method != req.Method {
			continue
		}

		matches := route.regex.FindStringSubmatch(req.URL.Path)
		if len(matches) > 0 {
			if len(matches) > 1 {
				params := make(map[string]string)
				for i, param := range route.params {
					params[param] = matches[i+1]
				}
				ctx := context.WithValue(req.Context(), PathParamsKey, params)
				req = req.WithContext(ctx)
			}
			route.handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

func (r *Router) HandleMethod(method, pattern string, handler http.HandlerFunc) {
	params := []string{}
	regexPattern := pattern

	if strings.Contains(pattern, "{") {
		paramPattern := regexp.MustCompile(`\{([^/]+)\}`)
		matches := paramPattern.FindAllStringSubmatch(pattern, -1)

		parts := paramPattern.Split(pattern, -1)
		for i, part := range parts {
			parts[i] = regexp.QuoteMeta(part)
		}

		regexPattern = parts[0]
		for i, match := range matches {
			params = append(params, match[1])
			regexPattern += "([^/]+)"
			if i+1 < len(parts) {
				regexPattern += parts[i+1]
			}
		}
	}

	r.routes = append(r.routes, &route{
		method:  method,
		regex:   regexp.MustCompile("^" + regexPattern + "$"),
		handler: handler,
		params:  params,
	})
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

func GetParam(r *http.Request, param string) string {
	params, ok := r.Context().Value(PathParamsKey).(map[string]string)
	if !ok {
		return ""
	}
	return params[param]
}