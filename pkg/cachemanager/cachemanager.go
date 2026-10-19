package cachemanager

import (
	"context"
	"fmt"
	"time"

	"github.com/gocasters/rankr/adapter/redis"
	goredis "github.com/redis/go-redis/v9"
)

type redisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *goredis.StatusCmd
	Get(ctx context.Context, key string) *goredis.StringCmd
	Del(ctx context.Context, keys ...string) *goredis.IntCmd
	TTL(ctx context.Context, key string) *goredis.DurationCmd
}

type CacheManager struct {
	client redisClient
}

func NewCacheManager(cache *redis.Adapter) *CacheManager {
	return &CacheManager{
		client: cache.Client(),
	}
}

func (c *CacheManager) Set(ctx context.Context, key string, value any, expire time.Duration) error {

	err := c.client.Set(ctx, key, value, expire).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheManager) Get(ctx context.Context, key string) (string, error) {
	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return data, nil
}

func (c *CacheManager) Delete(ctx context.Context, keys ...string) error {

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return err
	}

	return nil
}

func (c *CacheManager) GetTTL(ctx context.Context, key string) (int64, bool, error) {

	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, false, err
	}

	// -2s: key does not exist
	if ttl == -2*time.Second {
		return 0, false, fmt.Errorf("key does not exist")
	}

	// -1s: key exists with no associated expire
	if ttl == -1*time.Second {
		return -1, true, nil
	}

	return int64(ttl.Seconds()), true, nil
}
