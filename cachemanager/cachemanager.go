// cachemanager.go
package cachemanager

import (
	"context"
	"github.com/gocasters/rankr/adapter/redis"
	"time"
)

type CacheManager struct {
	cache *redis.Adapter
}

func NewCacheManager(cache *redis.Adapter) *CacheManager {
	return &CacheManager{
		cache: cache,
	}
}

func (c *CacheManager) Set(ctx context.Context, key string, value any, expire time.Duration) error {
	err := c.cache.Client().Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *CacheManager) Get(ctx context.Context, key string) (string, error) {
	data, err := c.cache.Client().Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return data, nil
}
