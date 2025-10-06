package natsadapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// StorageTypeEnum defines valid storage types
type StorageTypeEnum string

const (
	StorageMemory StorageTypeEnum = "memory"
	StorageFile   StorageTypeEnum = "file"
)

// ToNatsStorage converts StorageTypeEnum to nats.StorageType
func (s StorageTypeEnum) ToNatsStorage() nats.StorageType {
	switch s {
	case StorageMemory:
		return nats.MemoryStorage
	case StorageFile:
		return nats.FileStorage
	default:
		return nats.FileStorage // default fallback
	}
}

// RetentionPolicyEnum defines valid retention policies
type RetentionPolicyEnum string

const (
	RetentionLimits    RetentionPolicyEnum = "limits"
	RetentionWorkQueue RetentionPolicyEnum = "workqueue"
	RetentionInterest  RetentionPolicyEnum = "interest"
)

// ToNatsRetention converts RetentionPolicyEnum to nats.RetentionPolicy
func (r RetentionPolicyEnum) ToNatsRetention() nats.RetentionPolicy {
	switch r {
	case RetentionLimits:
		return nats.LimitsPolicy
	case RetentionWorkQueue:
		return nats.WorkQueuePolicy
	case RetentionInterest:
		return nats.InterestPolicy
	default:
		return nats.LimitsPolicy
	}
}

type Config struct {
	URL             string              `koanf:"url"`
	StreamName      string              `koanf:"stream_name"`
	MaxReconnects   int                 `koanf:"max_reconnects"`
	ReconnectWait   time.Duration       `koanf:"reconnect_wait"`
	StreamSubjects  []string            `koanf:"stream_subjects"`
	MaxMessages     int64               `koanf:"max_messages"`
	MaxBytes        int64               `koanf:"max_bytes"`
	MaxAge          time.Duration       `koanf:"max_age"`
	StorageType     StorageTypeEnum     `koanf:"storage_type"`
	Replicas        int                 `koanf:"replicas"`
	RetentionPolicy RetentionPolicyEnum `koanf:"retention_policy"`
}

type Adapter struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config Config
}

func New(config Config) (*Adapter, error) {
	opts := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				fmt.Printf("Disconnected from NATS: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("Reconnected to NATS: %s\n", nc.ConnectedUrl())
		}),
	}

	conn, err := nats.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("get JetStream context: %w", err)
	}

	adapter := &Adapter{
		conn:   conn,
		js:     js,
		config: config,
	}

	if err := adapter.ensureStream(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ensure stream: %w", err)
	}

	return adapter, nil
}

func (a *Adapter) ensureStream() error {
	streamConfig := &nats.StreamConfig{
		Name:      a.config.StreamName,
		Subjects:  a.config.StreamSubjects,
		MaxMsgs:   a.config.MaxMessages,
		MaxBytes:  a.config.MaxBytes,
		MaxAge:    a.config.MaxAge,
		Storage:   a.config.StorageType.ToNatsStorage(),
		Replicas:  a.config.Replicas,
		Retention: a.config.RetentionPolicy.ToNatsRetention(),
	}

	_, err := a.js.StreamInfo(a.config.StreamName)
	if err != nil {
		_, err = a.js.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("create stream: %w", err)
		}
		fmt.Printf("Created stream: %s\n", a.config.StreamName)
		return nil
	}

	_, err = a.js.UpdateStream(streamConfig)
	if err != nil {
		return fmt.Errorf("update stream: %w", err)
	}

	return nil
}

// Publish publishes a message to a subject
func (a *Adapter) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := a.js.Publish(subject, data, nats.Context(ctx))
	if err != nil {
		return fmt.Errorf("publish to subject %s: %w", subject, err)
	}

	return nil
}

// PublishAsync publishes a message asynchronously
func (a *Adapter) PublishAsync(subject string, data []byte) (nats.PubAckFuture, error) {
	future, err := a.js.PublishAsync(subject, data)
	if err != nil {
		return nil, fmt.Errorf("publish async to subject %s: %w", subject, err)
	}

	return future, nil
}

// PullConsumerConfig configuration for pull consumer
type PullConsumerConfig struct {
	Subject     string        `koanf:"subject"`
	DurableName string        `koanf:"durable_name"`
	BatchSize   int           `koanf:"batch_size"`
	MaxWait     time.Duration `koanf:"max_wait"`
	MaxDeliver  int           `koanf:"max_deliver"`
	AckWait     time.Duration `koanf:"ack_wait"`
}

// CreatePullConsumer creates a pull-based consumer
func (a *Adapter) CreatePullConsumer(config PullConsumerConfig) (*PullConsumer, error) {
	sub, err := a.js.PullSubscribe(
		config.Subject,
		config.DurableName,
		nats.BindStream(a.config.StreamName),
		nats.ManualAck(),
		nats.MaxDeliver(config.MaxDeliver),
		nats.AckWait(config.AckWait),
	)
	if err != nil {
		return nil, fmt.Errorf("create pull subscription: %w", err)
	}

	return &PullConsumer{
		sub:    sub,
		config: config,
	}, nil
}

// GetStreamInfo returns stream information
func (a *Adapter) GetStreamInfo() (*nats.StreamInfo, error) {
	return a.js.StreamInfo(a.config.StreamName)
}

// Close gracefully closes the connection
func (a *Adapter) Close() error {
	if a.conn != nil {
		a.conn.Close()
	}
	return nil
}

// PullConsumer represents a pull-based consumer
type PullConsumer struct {
	sub    *nats.Subscription
	config PullConsumerConfig
}

// Fetch retrieves a batch of messages
func (c *PullConsumer) Fetch() ([]*nats.Msg, error) {
	msgs, err := c.sub.Fetch(
		c.config.BatchSize,
		nats.MaxWait(c.config.MaxWait),
	)

	if err != nil {
		if errors.Is(err, nats.ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
			return nil, nil
		}
		return nil, fmt.Errorf("fetch messages: %w", err)
	}

	return msgs, nil
}

// GetConsumerInfo returns consumer statistics
func (c *PullConsumer) GetConsumerInfo() (*nats.ConsumerInfo, error) {
	return c.sub.ConsumerInfo()
}

// Close closes the subscription
func (c *PullConsumer) Close() error {
	return c.sub.Unsubscribe()
}
