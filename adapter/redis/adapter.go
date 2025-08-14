package redis

import (
    "github.com/redis/go-redis/v9"
)

// Adapter wraps the Redis client for our application
type Adapter struct {
    Client *redis.Client
}

// NewAdapter creates a new Redis adapter instance
func NewAdapter(client *redis.Client) *Adapter {
    return &Adapter{Client: client}
}
