package cache

import "github.com/redis/go-redis/v9"

type Config struct {
	RedisClient *redis.Client
}
