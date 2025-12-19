package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"github.com/gocasters/rankr/pkg/logger"
	redisv9 "github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
)

type BrokerConfig struct {
	StreamKey      string        `koanf:"stream_key"`
	GroupName      string        `koanf:"group_name"`
	ConsumerPrefix string        `koanf:"consumer_prefix"`
	BlockTime      time.Duration `koanf:"block_time"`
	BatchSize      int64         `koanf:"batch_size"`
	RetryCount     int           `koanf:"retry_count"`
}

type Broker struct {
	config BrokerConfig
	redis  *redis.Adapter
}

type Message struct {
	ID      string
	Payload []byte
	Retry   int
}

func NewBroker(cfg BrokerConfig, redis *redis.Adapter) Broker {
	return Broker{config: cfg, redis: redis}
}

func (b Broker) InitGroup(ctx context.Context) error {
	err := b.redis.Client().XGroupCreateMkStream(ctx, b.config.StreamKey, b.config.GroupName, "$").Err()
	if err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			return nil
		}

		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	return nil
}

func (b Broker) Publish(ctx context.Context, pj job.ProduceJob) error {
	idempotencyKey := fmt.Sprintf("job_id:%s", pj.IdempotencyKey)

	ok, err := b.redis.Client().SetNX(ctx, idempotencyKey, 1, time.Hour).Result()
	if err != nil {
		return fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !ok {
		return fmt.Errorf("job %s already published", pj.IdempotencyKey)
	}

	_, err = b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
		Stream: b.config.StreamKey,
		Values: map[string]interface{}{
			"job_id": pj.JobID,
			"retry":  0,
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to publish job to redis: %w", err)
	}

	return nil
}

func (b Broker) Ack(ctx context.Context, Id string) error {
	const maxAckRetry = 3
	var err error
	for i := 0; i < maxAckRetry; i++ {
		err = b.redis.Client().XAck(ctx, b.config.StreamKey, b.config.GroupName, Id).Err()
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("ack failed after %d retries: %w", maxAckRetry, err)
}

func (b Broker) Consume(ctx context.Context, consumer string) ([]Message, error) {
	res, err := b.redis.Client().XReadGroup(ctx, &redisv9.XReadGroupArgs{
		Group:    b.config.GroupName,
		Consumer: b.config.ConsumerPrefix + consumer,
		Streams:  []string{b.config.StreamKey, ">"},
		Count:    b.config.BatchSize,
		Block:    b.config.BlockTime,
	}).Result()

	if err != nil {
		if err == redisv9.Nil {
			return nil, nil
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		return nil, fmt.Errorf("read group failed: %w", err)
	}

	var msg []Message
	for _, stream := range res {
		for _, m := range stream.Messages {

			jobID, _ := m.Values["job_id"]
			retry, _ := strconv.Atoi(fmt.Sprint(m.Values["retry"]))

			if retry >= b.config.RetryCount {
				pErr := b.publishToDLQ(ctx, Message{
					ID:      m.ID,
					Payload: []byte(fmt.Sprint(jobID)),
					Retry:   retry,
				})

				if pErr != nil {
					logger.L().Error(
						"failed publish to DLQ",
						"message id",
						m.ID,
						"job_id",
						fmt.Sprint(jobID),
						"error",
						pErr.Error(),
					)
				}

				aErr := b.Ack(ctx, m.ID)

				if aErr != nil {
					logger.L().Error(
						"failed ack",
						"message id",
						m.ID,
						"job_id",
						fmt.Sprint(jobID),
						"error",
						aErr.Error(),
					)
				}

				continue
			}

			msg = append(msg, Message{
				ID:      m.ID,
				Payload: []byte(fmt.Sprint(jobID)),
				Retry:   retry,
			})
		}
	}

	return msg, nil
}

func (b Broker) HandleFailure(ctx context.Context, msg Message, procErr error) error {
	if msg.Retry >= b.config.RetryCount {
		if err := b.publishToDLQ(ctx, msg); err != nil {
			return fmt.Errorf("publish to DLQ failed: %w", err)
		}
		return b.Ack(ctx, msg.ID)
	}

	if err := b.requeue(ctx, msg); err != nil {
		return fmt.Errorf("requeue failed: %w", err)
	}

	return b.Ack(ctx, msg.ID)
}

func (b Broker) publishToDLQ(ctx context.Context, msg Message) error {
	dlqKey := fmt.Sprintf("DLQ:%s:%d", msg.ID, msg.Retry)

	ok, err := b.redis.Client().SetNX(ctx, dlqKey, 1, time.Hour).Result()
	if err != nil {
		return fmt.Errorf("failed to set DLQ key: %w", err)
	}
	if !ok {
		return nil
	}

	_, err = b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
		Stream: b.config.StreamKey + ".DLQ",
		Values: map[string]interface{}{
			"job_id":    string(msg.Payload),
			"retry":     msg.Retry,
			"dlq_key":   dlqKey,
			"failed_at": time.Now(),
		},
	}).Result()
	return err
}

func (b Broker) requeue(ctx context.Context, msg Message) error {
	_, err := b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
		Stream: b.config.StreamKey,
		Values: map[string]interface{}{
			"job_id": string(msg.Payload),
			"retry":  msg.Retry + 1,
		},
	}).Result()

	return err
}
