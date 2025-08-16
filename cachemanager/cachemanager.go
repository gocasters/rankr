// cachemanager.go
package cachemanager

import (
	"context"
	"rankr/adapter/redis"
	"time"
)

type CacheManager struct {
	cache *redis.Adapter
}

func NewCacheManager(cache *redis.Adapter) *CacheManager {
	if cache == nil
	{		
		return nil	
	}
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
// Get returns the cached value for a key. 
// It differentiates between a cache miss (found=false, err=nil) and real errors.
func (c *CacheManager) Get(ctx context.Context, key string) (string,bool, error) {
	data, err := c.cache.Client().Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false, nil

	}
	if err != nil {
		return "",false, err
	}
	return data,true, nil
}

func (c *CacheManager) Delete(ctx context.Context, key string) error {
    return c.store.Delete(ctx, key) 
}