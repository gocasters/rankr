package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/job"
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
	_, err := b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
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

			msg = append(msg, Message{
				ID:      m.ID,
				Payload: []byte(fmt.Sprint(jobID)),
				Retry:   retry,
			})
		}
	}

	return msg, nil
}

func (b Broker) Ack(ctx context.Context, ids ...string) error {
	if len(ids) < 1 {
		return nil
	}

	if err := b.redis.Client().XAck(ctx, b.config.StreamKey, b.config.GroupName, ids...).Err(); err != nil {
		return fmt.Errorf("ack failed: %w", err)
	}

	return nil
}

func (b Broker) HandleFailure(ctx context.Context, msg Message) error {
	if msg.Retry >= b.config.RetryCount {
		if err := b.publishToDLQ(ctx, msg); err != nil {
			return err
		}

		return b.Ack(ctx, msg.ID)
	}

	if err := b.requeue(ctx, msg); err != nil {
		return err
	}

	return b.Ack(ctx, msg.ID)
}

func (b Broker) publishToDLQ(ctx context.Context, msg Message) error {
	_, err := b.redis.Client().XAdd(ctx, &redisv9.XAddArgs{
		Stream: b.config.StreamKey + ".DLQ",
		Values: map[string]interface{}{
			"job_id":    string(msg.Payload),
			"retry":     msg.Retry,
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
