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
	errs := map[string]error{}
	if config.Host == "" {
		errs["host"] = fmt.Errorf("redis host is empty")
	}
	if config.Port <= 0 || config.Port > 65535 {
		errs["port"] = fmt.Errorf("invalid redis port: %d", config.Port)
	}
	// TODO: may be need to add maxDb in config or just remove the upper bound
	if config.DB < 0 || config.DB > 15 {
		errs["db"] = fmt.Errorf("invalid redis DB: %d, must be between 0 and 15", config.DB)
	}
	return errs
}

// "validation errors: field1: error1; field2: error2", joining each "field: error" pair with "; ".
func FormatValidationErrors(errs map[string]error) string {
	if len(errs) == 0 {
		return ""
	}

	var errorStrings []string
	for field, err := range errs {
		errorStrings = append(errorStrings, fmt.Sprintf("%s: %v", field, err))
	}

	return fmt.Sprintf("validation errors: %s", strings.Join(errorStrings, "; "))
}

type Adapter struct {
	client *redis.Client
}

// New creates a Redis Adapter from the supplied Config and verifies connectivity.
//
// It validates the configuration, constructs a redis.Client, and performs a Ping using
// the provided context to ensure the server is reachable. Returns an error if the
// context is nil, the configuration is invalid (error message includes formatted
// validation errors), or the connection ping fails (returned error wraps the ping error).
// On ping failure the created client is closed before returning.
func New(ctx context.Context, config Config) (*Adapter, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if validationErrors := config.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("invalid redis configuration: %s", FormatValidationErrors(validationErrors))
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Errorf("Failed to connect to Redis at %s (db=%d): %v", addr, config.DB, err)

		if cErr := rdb.Close(); cErr != nil {
			log.Errorf("Error closing Redis client after connection failure: %v", cErr)
		}

		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Infof("✅ Redis is up and running at %s (db=%d)", addr, config.DB)

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
