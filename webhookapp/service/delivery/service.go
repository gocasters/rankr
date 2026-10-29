package delivery

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type EventRepository interface {
	Save(ctx context.Context, event *eventpb.Event) error
	BulkInsertPostgresSQL(ctx context.Context, events []string) ([]*eventpb.Event, error)
	GetLostDeliveries(ctx context.Context, provider eventpb.EventProvider, deliveries []string) ([]string, error)
}
type EventDurableRepository interface {
	GetRedisClient() *redis.Client
	GetBatchFromRedis(ctx context.Context, queueName string, batchSize int64) ([]string, error)
	RequeueFailedEvents(ctx context.Context, queueName string, events []string)
}

type Service struct {
	repo               EventRepository
	publisher          message.Publisher
	durableRepo        EventDurableRepository
	githubHmacSecret   []byte
	githubMaxEventSize int64
	insertQueueName    string
	insertBatchSize    int64
}

func New(repo EventRepository, publisher message.Publisher, durableRepo EventDurableRepository, githubHmacSecret []byte, githubMaxEventSize int64, insertQueueName string, insertBatchSize int64) *Service {
	return &Service{
		repo:               repo,
		publisher:          publisher,
		durableRepo:        durableRepo,
		githubHmacSecret:   githubHmacSecret,
		githubMaxEventSize: githubMaxEventSize,
		insertQueueName:    insertQueueName,
		insertBatchSize:    insertBatchSize,
	}
}

func (s *Service) validateGitHubHeaders(hookID, eventName, deliveryUID, msgSig256 string) error {
	if hookID == "" {
		return fmt.Errorf("missing X-GitHub-Hook-ID header")
	}
	if eventName == "" {
		return fmt.Errorf("missing X-GitHub-Event header")
	}
	if deliveryUID == "" {
		return fmt.Errorf("missing X-GitHub-Delivery header")
	}
	if msgSig256 == "" {
		return fmt.Errorf("missing X-Hub-Signature-256 header")
	}
	return nil
}

func (s *Service) saveEvent(ctx context.Context, event *eventpb.Event) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = s.durableRepo.GetRedisClient().RPush(ctx, s.insertQueueName, payload).Result()
	if err != nil {
		return fmt.Errorf("failed to push to Redis: %w", err)
	}

	queueLength, err := s.durableRepo.GetRedisClient().LLen(ctx, s.insertQueueName).Result()
	if err != nil {
		return fmt.Errorf("failed to get queue length: %w", err)
	}

	if queueLength >= s.insertBatchSize {
		go func() {
			err := s.ProcessBatch(context.Background())
			if err != nil {
				// log error
			}
		}()
	}

	return nil
}

func (s *Service) VerifyEvent(eventReader io.Reader, hookID, eventName, deliveryUID, hubSig256 string) ([]byte, error) {
	if len(s.githubHmacSecret) == 0 {
		return nil, fmt.Errorf("GitHub webhook secret must be set")
	}
	if s.githubMaxEventSize <= 0 {
		return nil, fmt.Errorf("GitHub payload size limit must be positive")
	}

	if err := s.validateGitHubHeaders(hookID, eventName, deliveryUID, hubSig256); err != nil {
		return nil, err
	}

	event, err := io.ReadAll(io.LimitReader(eventReader, s.githubMaxEventSize+1))
	if err != nil {
		return nil, fmt.Errorf("read GitHub payload: %w", err)
	}
	if int64(len(event)) > s.githubMaxEventSize {
		return nil, fmt.Errorf("GitHub payload exceeds the %d-byte limit", s.githubMaxEventSize)
	}

	const ghSigPrefix = "sha256="
	if !strings.HasPrefix(hubSig256, ghSigPrefix) {
		return nil, fmt.Errorf("GitHub signature must start with %q", ghSigPrefix)
	}

	decodedSig, err := hex.DecodeString(strings.TrimPrefix(hubSig256, ghSigPrefix))
	if err != nil {
		return nil, fmt.Errorf("decode GitHub signature: %w", err)
	}
	mac := hmac.New(sha256.New, s.githubHmacSecret)
	_, _ = mac.Write(event)
	if !hmac.Equal(mac.Sum(nil), decodedSig) {
		return nil, fmt.Errorf("hmac verification failed")
	}

	return event, nil
}

func (s *Service) ProcessBatch(ctx context.Context) error {
	events, err := s.durableRepo.GetBatchFromRedis(ctx, s.insertQueueName, s.insertBatchSize)
	if err != nil {
		return fmt.Errorf("failed to get batch from Redis: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	logger.L().Debug("row events", slog.Any("events", events))

	savedEvents, err := s.repo.BulkInsertPostgresSQL(ctx, events)
	if err != nil {
		s.durableRepo.RequeueFailedEvents(ctx, s.insertQueueName, events)
		return fmt.Errorf("bulk insert failed: %w", err)
	}

	logger.L().Debug("events", slog.Any("savedEvents", savedEvents))

	for _, ev := range savedEvents {
		payload, err := proto.Marshal(ev)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf event. eventname: %s. error: %w",
				ev.EventName, err)
		}

		msg := message.NewMessage(watermill.NewUUID(), payload)

		if err := s.publisher.Publish("rankr_raw_events", msg); err != nil {
			return fmt.Errorf("failed to publish event. eventname: %s, error: %w",
				ev.EventName, err)
		}
	}

	return nil
}

func (s *Service) GetLostDeliveries(ctx context.Context, provider eventpb.EventProvider, deliveries []string) ([]string, error) {
	return s.repo.GetLostDeliveries(ctx, provider, deliveries)
}
