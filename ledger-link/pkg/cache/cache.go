package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	RedisClient *redis.Client
}

const (
	ShortTerm  = 5 * time.Minute
	MediumTerm = 30 * time.Minute
	LongTerm   = 24 * time.Hour

	KeyBalance     = "balance"
	KeyUser        = "user"
	KeyTransaction = "transaction"
)

func NewCacheService(redisClient *redis.Client) *CacheService {
	return &CacheService{
		RedisClient: redisClient,
	}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.RedisClient.Set(ctx, key, data, expiration).Err()
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.RedisClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.RedisClient.Del(ctx, key).Err()
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}

func BuildKey(entity string, id uint) string {
	return fmt.Sprintf("%s:%d", entity, id)
}
