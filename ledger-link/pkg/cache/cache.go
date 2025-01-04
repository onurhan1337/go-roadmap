package cache

import "github.com/redis/go-redis/v9"

type CacheService struct {
	redis *redis.Client
}
