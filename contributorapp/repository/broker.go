package repository

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/job"
	redisv9 "github.com/redis/go-redis/v9"
)

type BrokerConfig struct {
	StreamKey string `koanf:"stream_key"`
}

type Broker struct {
	config BrokerConfig
	redis  *redis.Adapter
}

func NewBroker(cfg BrokerConfig, redis *redis.Adapter) Broker {
	return Broker{config: cfg, redis: redis}
}

func (b Broker) Publish(ctx context.Context, pj job.ProduceJob) error {
	_, err := b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
		Stream: b.config.StreamKey,
		Values: map[string]interface{}{
			"job_id":    pj.JobID,
			"file_path": pj.FilePath,
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to publish job to redis: %w", err)
	}

	return nil
}
