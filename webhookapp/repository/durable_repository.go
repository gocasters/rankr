package repository

import (
	"context"
	"fmt"
	adapterRedis "github.com/gocasters/rankr/adapter/redis"
	"github.com/redis/go-redis/v9"
)

type WebhookDurableRepository struct {
	redisAdapter *adapterRedis.Adapter
}

func NewWebhookDurableRepository(redisAdapter *adapterRedis.Adapter) WebhookDurableRepository {
	return WebhookDurableRepository{redisAdapter: redisAdapter}
}

func (repo *WebhookDurableRepository) GetRedisClient() *redis.Client {
	return repo.redisAdapter.Client()
}

func (repo *WebhookDurableRepository) GetBatchFromRedis(ctx context.Context, queueName string, batchSize int64) ([]string, error) {
	pipe := repo.redisAdapter.Client().TxPipeline()

	eventsCmd := pipe.LRange(ctx, queueName, 0, batchSize-1)
	pipe.LTrim(ctx, queueName, batchSize, -1)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis transaction failed: %w", err)
	}

	return eventsCmd.Val(), nil
}

func (repo *WebhookDurableRepository) RequeueFailedEvents(ctx context.Context, queueName string, events []string) {
	for _, event := range events {
		repo.redisAdapter.Client().LPush(ctx, queueName, event)
	}
}
