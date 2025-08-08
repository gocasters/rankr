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
		log.Errorf("Failed to connect to Redis: %v", err)

		if cErr := rdb.Close(); cErr != nil {
			log.Errorf("Error closing Redis client after connection failure: %v", cErr)
		}

		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Info("✅ Redis is up running...")

	return &Adapter{client: rdb}, nil
}

func (a Adapter) Client() *redis.Client {
	return a.client
}
