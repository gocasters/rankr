package worker

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/contributorapp/repository"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Config struct {
	WorkerCount int `koanf:"worker_count"`
}

type Consumer interface {
	Consume(ctx context.Context, consumer string) ([]repository.Message, error)
	Ack(ctx context.Context, ids ...string) error
	HandleFailure(ctx context.Context, msg repository.Message) error
}

type Pool struct {
	consumer Consumer
	worker   Worker
	config   Config
}

func New(consumer Consumer, worker Worker, cfg Config) Pool {
	return Pool{consumer: consumer, worker: worker, config: cfg}
}

func (p Pool) Start(ctx context.Context) {
	wg := &sync.WaitGroup{}
	msgCh := make(chan repository.Message, p.config.WorkerCount)

	for i := 1; i <= p.config.WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for msg := range msgCh {
				if err := p.worker.Process(ctx, string(msg.Payload)); err != nil {
					if fErr := p.consumer.HandleFailure(ctx, msg); fErr != nil {
						logger.L().Error(
							"failed to handle failed job",
							"job_id", string(msg.Payload),
							"error", fErr)
					}

					continue
				}

				if err := p.consumer.Ack(ctx, msg.ID); err != nil {
					logger.L().Error(fmt.Sprintf("failed to ack message, id: %s", msg.ID),
						"error", err)
				}
			}
		}()
	}

	consumerName := fmt.Sprintf("%s", uuid.NewString())
	for {
		select {
		case <-ctx.Done():
			close(msgCh)
			wg.Wait()
			return
		default:
			msgs, err := p.consumer.Consume(ctx, consumerName)
			if err != nil {
				logger.L().Error("consumer error", "error", err.Error())
				time.Sleep(time.Second)
				continue
			}

			if len(msgs) == 0 {
				time.Sleep(100 * time.Millisecond) // avoid spinning when queue is empty
				continue
			}

			for _, m := range msgs {
				msgCh <- m
			}
		}
	}
}
