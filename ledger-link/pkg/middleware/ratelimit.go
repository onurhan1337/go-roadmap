package middleware

import (
	"ledger-link/pkg/ratelimit"
	"net/http"
	"time"
)

var (
	LoginRateLimit = ratelimit.RateLimit{
		Limit:    5,
		Duration: time.Minute,
	}
	RegisterRateLimit = ratelimit.RateLimit{
		Limit:    3,
		Duration: time.Minute,
	}
	TransactionRateLimit = ratelimit.RateLimit{
		Limit:    10,
		Duration: time.Minute,
	}
	BalanceOperationRateLimit = ratelimit.RateLimit{
		Limit:    20,
		Duration: time.Minute,
	}
	UserOperationRateLimit = ratelimit.RateLimit{
		Limit:    30,
		Duration: time.Minute,
	}
)

type RateLimitMiddleware struct {
	limiter *ratelimit.RateLimiter
}

func NewRateLimitMiddleware(limiter *ratelimit.RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimitMiddleware) LoginLimit(next http.Handler) http.Handler {
	return m.limiter.Limit("login", LoginRateLimit)(next)
}

func (m *RateLimitMiddleware) RegisterLimit(next http.Handler) http.Handler {
	return m.limiter.Limit("register", RegisterRateLimit)(next)
}

func (m *RateLimitMiddleware) TransactionLimit(next http.Handler) http.Handler {
	return m.limiter.Limit("transaction", TransactionRateLimit)(next)
}

func (m *RateLimitMiddleware) BalanceLimit(next http.Handler) http.Handler {
	return m.limiter.Limit("balance", BalanceOperationRateLimit)(next)
}

func (m *RateLimitMiddleware) UserOperationLimit(next http.Handler) http.Handler {
	return m.limiter.Limit("user", UserOperationRateLimit)(next)
}
