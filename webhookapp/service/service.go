package service

import (
	"context"
	"fmt"
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
}
type EventDurableRepository interface {
	GetRedisClient() *redis.Client
	GetBatchFromRedis(ctx context.Context, queueName string, batchSize int64) ([]string, error)
	RequeueFailedEvents(ctx context.Context, queueName string, events []string)
}

type Service struct {
	repo            EventRepository
	publisher       message.Publisher
	durableRepo     EventDurableRepository
	insertQueueName string
	insertBatchSize int64
}

func New(repo EventRepository, publisher message.Publisher, durableRepo EventDurableRepository, insertQueueName string, insertBatchSize int64) *Service {
	return &Service{
		repo:            repo,
		publisher:       publisher,
		durableRepo:     durableRepo,
		insertQueueName: insertQueueName,
		insertBatchSize: insertBatchSize,
	}
}

func (s *Service) publishEvent(ev *eventpb.Event, evName eventpb.EventName, topic Topic, metadata map[string]string) error {
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
