package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

func (config Config) Validate() map[string]error {
	errors := map[string]error{}
	if config.Host == "" {
		errors["host"] = fmt.Errorf("redis host is empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		errors["port"] = fmt.Errorf("invalid redis port: %d", config.Port)
	}
	if config.DB < 0 || config.DB > 15 {
		errors["db"] = fmt.Errorf("invalid redis DB: %d, must be between 0 and 15", config.DB)
	}
	return errors
}

func FormatValidationErrors(errors map[string]error) string {
	if len(errors) == 0 {
		return ""
	}

	var errorStrings []string
	for field, err := range errors {
		errorStrings = append(errorStrings, fmt.Sprintf("%s: %v", field, err))
	}

	return fmt.Sprintf("validation errors: %s", strings.Join(errorStrings, "; "))
}

type Adapter struct {
	client *redis.Client
}

func New(ctx context.Context, config Config) (*Adapter, error) {
	if validationErrors := config.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("invalid redis configuration: %s", FormatValidationErrors(validationErrors))
	}

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

func (a *Adapter) Client() *redis.Client {
	return a.client
}

func (a *Adapter) Close() error {
	if a == nil || a.client == nil {
		return nil
	}

	return a.client.Close()
}
