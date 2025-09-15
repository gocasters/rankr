package nats

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill/message"
	natsgo "github.com/nats-io/nats.go"
)

// Config represents NATS configuration
type Config struct {
	URL            string        `koanf:"url"`
	ClusterID      string        `koanf:"cluster_id"`
	ClientID       string        `koanf:"client_id"`
	DurableName    string        `koanf:"durable_name"`
	QueueGroup     string        `koanf:"queue_group"`
	AckWaitTimeout time.Duration `koanf:"ack_wait_timeout"`
	MaxInflight    int           `koanf:"max_inflight"`
	ConnectTimeout time.Duration `koanf:"connect_timeout"`
	ReconnectWait  time.Duration `koanf:"reconnect_wait"`
	MaxReconnects  int           `koanf:"max_reconnects"`
	PingInterval   time.Duration `koanf:"ping_interval"`
	MaxPingsOut    int           `koanf:"max_pings_out"`
	AllowReconnect bool          `koanf:"allow_reconnect"`
	UseJetStream   bool          `koanf:"use_jetstream"`
}

// Validate validates the NATS configuration
func (c Config) Validate() map[string]error {
	errs := map[string]error{}

	if strings.TrimSpace(c.URL) == "" {
		errs["url"] = fmt.Errorf("NATS URL cannot be empty")
	}

	if strings.TrimSpace(c.ClientID) == "" {
		errs["client_id"] = fmt.Errorf("client ID cannot be empty")
	}

	if c.MaxInflight <= 0 {
		errs["max_inflight"] = fmt.Errorf("max inflight must be positive, got: %d", c.MaxInflight)
	}

	if c.MaxReconnects < -1 || c.MaxReconnects == 0 {
		errs["max_reconnects"] = fmt.Errorf("max reconnects must be -1 (unlimited) or positive (0 is not allowed, use -1 for unlimited), got: %d", c.MaxReconnects)
	}

	if c.PingInterval <= 0 {
		errs["ping_interval"] = fmt.Errorf("ping interval must be positive, got: %v", c.PingInterval)
	}

	if c.MaxPingsOut <= 0 {
		errs["max_pings_out"] = fmt.Errorf("max pings out must be positive, got: %d", c.MaxPingsOut)
	}

	return errs
}

// SetDefaults sets default values for the configuration
func (c *Config) SetDefaults() {
	if c.URL == "" {
		c.URL = natsgo.DefaultURL
	}
	if c.AckWaitTimeout == 0 {
		c.AckWaitTimeout = 30 * time.Second
	}
	if c.MaxInflight == 0 {
		c.MaxInflight = 1024
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 5 * time.Second
	}
	if c.ReconnectWait == 0 {
		c.ReconnectWait = 2 * time.Second
	}
	if c.MaxReconnects == 0 {
		c.MaxReconnects = -1 // unlimited
	}
	if c.PingInterval == 0 {
		c.PingInterval = 2 * time.Minute
	}
	if c.MaxPingsOut == 0 {
		c.MaxPingsOut = 2
	}
	c.AllowReconnect = true // default to true
}

// FormatValidationErrors formats validation errors into a readable string
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

// Adapter wraps NATS publisher and subscriber with Watermill integration
type Adapter struct {
	config     Config
	publisher  message.Publisher
	subscriber message.Subscriber
	conn       *natsgo.Conn
	logger     watermill.LoggerAdapter
}

// New creates a new NATS adapter with Watermill integration
func New(ctx context.Context, config Config, logger watermill.LoggerAdapter) (adapter *Adapter, initErr error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}

	if logger == nil {
		logger = watermill.NopLogger{}
	}

	config.SetDefaults()
	if validationErrors := config.Validate(); len(validationErrors) > 0 {
		return nil, fmt.Errorf("invalid NATS configuration: %s", FormatValidationErrors(validationErrors))
	}

	opts := []natsgo.Option{
		natsgo.Timeout(config.ConnectTimeout),
		natsgo.ReconnectWait(config.ReconnectWait),
		natsgo.MaxReconnects(config.MaxReconnects),
		natsgo.PingInterval(config.PingInterval),
		natsgo.MaxPingsOutstanding(config.MaxPingsOut),
	}

	if !config.AllowReconnect {
		opts = append(opts, natsgo.NoReconnect())
	}

	conn, err := natsgo.Connect(config.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS at %s: %w", config.URL, err)
	}

	defer func() {
		if initErr != nil && conn != nil {
			conn.Close()
		}
	}()

	var publisher message.Publisher
	var subscriber message.Subscriber

	marshaler := &nats.NATSMarshaler{}

	if config.UseJetStream {

		jsConfig := nats.JetStreamConfig{
			Disabled:      false,
			AutoProvision: true,
			SubscribeOptions: []natsgo.SubOpt{
				natsgo.DeliverAll(),
				natsgo.AckExplicit(),
				natsgo.AckWait(config.AckWaitTimeout),
				natsgo.MaxDeliver(10),
			},
			AckAsync:      false,
			DurablePrefix: config.DurableName,
		}

		publisherConfig := nats.PublisherConfig{
			URL:         config.URL,
			NatsOptions: opts,
			Marshaler:   marshaler,
			JetStream:   jsConfig,
		}

		publisher, err = nats.NewPublisher(publisherConfig, logger)
		if err != nil {
			initErr = fmt.Errorf("failed to create JetStream publisher: %w", err)
			return
		}

		subscriberConfig := nats.SubscriberConfig{
			URL:            config.URL,
			NatsOptions:    opts,
			Unmarshaler:    marshaler,
			AckWaitTimeout: config.AckWaitTimeout,
			JetStream:      jsConfig,
		}

		subscriber, err = nats.NewSubscriber(subscriberConfig, logger)
		if err != nil {
			initErr = fmt.Errorf("failed to create JetStream subscriber: %w", err)
			return
		}
	} else {
		publisherConfig := nats.PublisherConfig{
			URL:         config.URL,
			NatsOptions: opts,
			Marshaler:   marshaler,
		}

		publisher, err = nats.NewPublisher(publisherConfig, logger)
		if err != nil {
			initErr = fmt.Errorf("failed to create NATS publisher: %w", err)
			return
		}

		subscriberConfig := nats.SubscriberConfig{
			URL:            config.URL,
			NatsOptions:    opts,
			Unmarshaler:    marshaler,
			AckWaitTimeout: config.AckWaitTimeout,
		}

		subscriber, err = nats.NewSubscriber(subscriberConfig, logger)
		if err != nil {
			initErr = fmt.Errorf("failed to create NATS subscriber: %w", err)
			return
		}
	}

	if err := conn.Flush(); err != nil {
		initErr = fmt.Errorf("failed to flush NATS connection: %w", err)
		return
	}

	adapter = &Adapter{
		config:     config,
		publisher:  publisher,
		subscriber: subscriber,
		conn:       conn,
		logger:     logger,
	}

	logger.Info("NATS adapter successfully connected",
		watermill.LogFields{
			"url":       config.URL,
			"client_id": config.ClientID,
			"jetstream": config.UseJetStream,
		})

	return adapter, nil
}

// Publisher returns the Watermill message publisher
func (a *Adapter) Publisher() message.Publisher {
	return a.publisher
}

// Subscriber returns the Watermill message subscriber
func (a *Adapter) Subscriber() message.Subscriber {
	return a.subscriber
}

// Connection returns the underlying NATS connection
func (a *Adapter) Connection() *natsgo.Conn {
	return a.conn
}

// Config returns the adapter configuration
func (a *Adapter) Config() Config {
	return a.config
}

// IsConnected checks if the NATS connection is still active
func (a *Adapter) IsConnected() bool {
	return a.conn != nil && a.conn.IsConnected()
}

// Status returns the current connection status
func (a *Adapter) Status() natsgo.Status {
	if a.conn == nil {
		return natsgo.DISCONNECTED
	}
	return a.conn.Status()
}

// Publish publishes a message to the specified topic
func (a *Adapter) Publish(topic string, msg *message.Message) error {
	if a.publisher == nil {
		return fmt.Errorf("publisher is not initialized")
	}
	return a.publisher.Publish(topic, msg)
}

// Subscribe subscribes to messages from the specified topic
func (a *Adapter) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	if a.subscriber == nil {
		return nil, fmt.Errorf("subscriber is not initialized")
	}
	return a.subscriber.Subscribe(ctx, topic)
}

// Close gracefully closes the NATS adapter and all its connections
func (a *Adapter) Close() error {
	var errs []error

	if a.publisher != nil {
		if err := a.publisher.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close publisher: %w", err))
		}
	}

	if a.subscriber != nil {
		if err := a.subscriber.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close subscriber: %w", err))
		}
	}

	if a.conn != nil {
		a.conn.Close()
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during NATS adapter shutdown: %v", errs)
	}

	a.logger.Info("NATS adapter closed successfully", nil)
	return nil
}

// Flush ensures all published messages have been sent
func (a *Adapter) Flush() error {
	if a.conn == nil {
		return fmt.Errorf("NATS connection is nil")
	}
	return a.conn.Flush()
}

// FlushTimeout ensures all published messages have been sent within the timeout
func (a *Adapter) FlushTimeout(timeout time.Duration) error {
	if a.conn == nil {
		return fmt.Errorf("NATS connection is nil")
	}
	return a.conn.FlushTimeout(timeout)
}
