package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/redis/go-redis/v9"
)

type cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) port.Cache {
	return &cache{rdb: rdb}
}

func (c *cache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}

func (c *cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := c.rdb.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *cache) Delete(ctx context.Context, key string) error {
	if err := c.rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}
