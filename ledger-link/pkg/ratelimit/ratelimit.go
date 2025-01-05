package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
}

type RateLimit struct {
	Limit    int
	Duration time.Duration
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
	}
}

func (rl *RateLimiter) Limit(key string, limit RateLimit) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identifier := getClientIdentifier(r)
			redisKey := fmt.Sprintf("ratelimit:%s:%s", key, identifier)

			count, err := rl.redisClient.Get(context.Background(), redisKey).Int()
			if err != nil && err != redis.Nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if count >= limit.Limit {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			pipe := rl.redisClient.Pipeline()
			pipe.Incr(context.Background(), redisKey)
			if count == 0 {
				pipe.Expire(context.Background(), redisKey, limit.Duration)
			}
			_, err = pipe.Exec(context.Background())
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", limit.Limit-count-1))

			next.ServeHTTP(w, r)
		})
	}
}

func getClientIdentifier(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	return r.RemoteAddr
}
