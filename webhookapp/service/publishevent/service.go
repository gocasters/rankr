package publishevent

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type EventRepository interface {
	Save(ctx context.Context, event *eventpb.Event) error
	BulkInsertPostgresSQL(ctx context.Context, events []string) ([]*eventpb.Event, error)
}
type EventDurableRepository interface {
	GetRedisClient() *redis.Client
	GetBatchFromRedis(ctx context.Context, queueName string, batchSize int64) ([]string, error)
	RequeueFailedEvents(ctx context.Context, queueName string, events []string)
}

type RawEventRepository interface {
	Save(ctx context.Context, event *service.RawEvent) error
}

type Service struct {
	repo            EventRepository
	rawRepo         RawEventRepository
	publisher       message.Publisher
	durableRepo     EventDurableRepository
	insertQueueName string
	insertBatchSize int64
}

// New creates a Service that persists events using repo and publishes them with publisher.
// The returned *Service coordinates saving events via the provided EventRepository and RawEventRepository and
// publishing Watermill messages using the provided message.Publisher.
func New(repo EventRepository, rawRepo RawEventRepository, publisher message.Publisher, durableRepo EventDurableRepository, insertQueueName string, insertBatchSize int64) *Service {
	return &Service{
		repo:            repo,
		rawRepo:         rawRepo,
		publisher:       publisher,
		durableRepo:     durableRepo,
		insertQueueName: insertQueueName,
		insertBatchSize: insertBatchSize,
	}
}

func (s *Service) publishEvent(ev *eventpb.Event, evName eventpb.EventName, topic service.Topic, metadata map[string]string) error {
	sErr := s.repo.Save(context.Background(), ev)
	if sErr != nil {
		return fmt.Errorf("failed to save event. eventname: %s, error: %w",
			evName, sErr)
	}

	payload, err := proto.Marshal(ev)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf event. eventname: %s. error: %w",
			evName, err)
	}

	msg := message.NewMessage(watermill.NewUUID(), payload)
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	fmt.Printf("event %s published to %s\n", evName, topic)

	if err := s.publisher.Publish(string(topic), msg); err != nil {
		return fmt.Errorf("failed to publish event. topic: %s, eventname: %s, error: %w",
			topic, evName, err)
	}

	return nil
}

func (s *Service) SaveRawEvent(event *service.RawEvent) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if err := s.rawRepo.Save(context.Background(), event); err != nil {
		return fmt.Errorf("failed to save raw event. provider: %s, deliveryID: %s, error: %w",
			event.Provider, event.DeliveryID, err)
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

func (s *Service) ProcessBatch(ctx context.Context) error {
	events, err := s.durableRepo.GetBatchFromRedis(ctx, s.insertQueueName, s.insertBatchSize)
	if err != nil {
		return fmt.Errorf("failed to get batch from Redis: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	logger.L().Debug("row events", events)

	savedEvents, err := s.repo.BulkInsertPostgresSQL(ctx, events)
	if err != nil {
		s.durableRepo.RequeueFailedEvents(ctx, s.insertQueueName, events)
		return fmt.Errorf("bulk insert failed: %w", err)
	}

	logger.L().Debug("events", savedEvents)

	for _, ev := range savedEvents {
		payload, err := proto.Marshal(ev)
		if err != nil {
			return fmt.Errorf("failed to marshal protobuf event. eventname: %s. error: %w",
				ev.EventName, err)
		}

		msg := message.NewMessage(watermill.NewUUID(), payload)

		// todo add helper to get each event topic and replace with ev.EventName
		if err := s.publisher.Publish(string(ev.EventName), msg); err != nil {
			return fmt.Errorf("failed to publish event. eventname: %s, error: %w",
				ev.EventName, err)
		}
	}

	return nil
}
