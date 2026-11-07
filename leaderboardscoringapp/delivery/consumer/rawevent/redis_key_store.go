package rawevent

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisKeyStore struct {
	client *redis.Client
}

func (s redisKeyStore) Exists(ctx context.Context, key string) (bool, error) {
	n, err := s.client.Exists(ctx, key).Result()
	return n == 1, err
}

func (s redisKeyStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

func (s redisKeyStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s redisKeyStore) DeleteIfEqual(ctx context.Context, key, value string) error {
	err := s.client.Watch(ctx, func(tx *redis.Tx) error {
		current, err := tx.Get(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			return nil
		}
		if err != nil {
			return err
		}
		if current != value {
			return nil
		}
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, key)
			return nil
		})
		return err
	}, key)
	if errors.Is(err, redis.TxFailedErr) {
		return nil
	}
	return err
}
