package rawevent

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var releaseLockLua = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
`)

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
	return releaseLockLua.Run(ctx, s.client, []string{key}, value).Err()
}
