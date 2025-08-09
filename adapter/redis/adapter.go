package redis

import (
	"context"
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type Adapter struct {
	client *redis.Client
}

func New(ctx context.Context, config Config) (*Adapter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Errorf("Failed to connect to Redis at %s:%d (db=%d): %v", config.Host, config.Port, config.DB, err)

		if cErr := rdb.Close(); cErr != nil {
			log.Errorf("Error closing Redis client after connection failure: %v", cErr)
		}

		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Infof("✅ Redis is up and running at %s:%d (db=%d)", config.Host, config.Port, config.DB)

	return &Adapter{client: rdb}, nil
}

func (a Adapter) Client() *redis.Client {
	return a.client
}

func (a Adapter) Close() error {
	if a.client == nil {
		return nil
	}

	return a.client.Close()
}
