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

func (c *Config) SetDefaults() {
	if c.MaxReconnects == 0 {
		c.MaxReconnects = 10
	}
	if c.ReconnectWait == 0 {
		c.ReconnectWait = 2 * time.Second
	}
	if c.Replicas == 0 {
		c.Replicas = 1
	}
	if c.StorageType == "" {
		c.StorageType = StorageFile
	}
	if c.RetentionPolicy == "" {
		c.RetentionPolicy = RetentionLimits
	}
}

func (c *Config) Validate() []string {
	var errors []string
	if c.URL == "" {
		errors = append(errors, "URL is required")
	}
	if c.StreamName == "" {
		errors = append(errors, "StreamName is required")
	}
	if len(c.StreamSubjects) == 0 {
		errors = append(errors, "StreamSubjects cannot be empty")
	}
	if c.MaxMessages < 0 {
		errors = append(errors, "MaxMessages cannot be negative")
	}
	if c.MaxBytes < 0 {
		errors = append(errors, "MaxBytes cannot be negative")
	}
	return errors
}

type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

type Adapter struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config Config
	logger Logger
}

func New(config Config, logger Logger) (*Adapter, error) {
	config.SetDefaults()
	if errs := config.Validate(); len(errs) > 0 {
		return nil, fmt.Errorf("invalid config: %v", errs)
	}

	opts := []nats.Option{
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(config.MaxReconnects),
		nats.ReconnectWait(config.ReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				logger.Error("Disconnected from NATS", "error", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("Reconnected to NATS", "url", nc.ConnectedUrl())
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
		logger: logger,
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

	fmt.Println(a.config.StreamName)
	fmt.Println(a.config.StreamSubjects)
	_, err := a.js.StreamInfo(a.config.StreamName)
	if err != nil {
		_, err = a.js.AddStream(streamConfig)
		if err != nil {
			return fmt.Errorf("create stream: %w", err)
		}
		a.logger.Info("Created stream", "name", a.config.StreamName)
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
		nats.AckExplicit(),
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

// CanRetry determines whether a message is still eligible for redelivery (retry)
// based on its delivery count and the configured MaxDeliver setting.
//
// This function returns `true` if the message can still be retried,
// or `false` if it has reached the delivery limit and should be sent to DLQ.
func (c *PullConsumer) CanRetry(msg *nats.Msg) bool {
	meta, _ := msg.Metadata()
	if c.config.MaxDeliver == int(meta.NumDelivered) {
		return false
	}
	return true
}

// GetConsumerInfo returns consumer statistics
func (c *PullConsumer) GetConsumerInfo() (*nats.ConsumerInfo, error) {
	return c.sub.ConsumerInfo()
}

// Close closes the subscription
func (c *PullConsumer) Close() error {
	return c.sub.Unsubscribe()
}
