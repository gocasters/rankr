package main

import (
	"context"
	"encoding/json"
	"fmt"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

// Global variables
var (
	redisBulkInserter *RedisBulkInserter
	pgPool            *pgxpool.Pool
	redisClient       *redis.Client
	totalProcessed    atomic.Int64
)

func init() {
	if err := initializeServices(); err != nil {
		log.Fatalf("failed to initialize services: %v", err)
	}
}

func initializeServices() error {
	var err error

	// Init Redis
	redisClient, err = initRedis()
	if err != nil {
		return fmt.Errorf("redis init failed: %w", err)
	}

	// Init Postgres
	pgPool, err = initPostgres()
	if err != nil {
		return fmt.Errorf("postgres init failed: %w", err)
	}

	// Init Bulk Inserter
	redisBulkInserter = NewRedisBulkInserter(redisClient, pgPool, 1000)

	// Start worker
	ctx := context.Background()
	go redisBulkInserter.StartWorker(ctx)

	return nil
}

func initRedis() (*redis.Client, error) {
	addr := getEnv("REDIS_ADDR", "webhook-redis:6379")
	password := getEnv("REDIS_PASSWORD", "")
	db := 0

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 100,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func initPostgres() (*pgxpool.Pool, error) {
	dsn := getEnv(
		"POSTGRES_URL",
		"postgres://webhook_admin:password123@webhook-db:5432/webhook_db",
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.MaxConns = 50

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMain(m *testing.M) {
	code := m.Run()

	cleanup()

	os.Exit(code)
}

func cleanup() {
	ctx := context.Background()

	// Truncate Postgres table
	_, err := pgPool.Exec(ctx, "TRUNCATE TABLE webhook_events RESTART IDENTITY CASCADE;")
	if err != nil {
		log.Printf("failed to truncate Postgres table: %v", err)
	} else {
		log.Println("✅ Postgres table truncated")
	}

	// Clear Redis queue
	_, err = redisClient.Del(ctx, "webhook_events").Result()
	if err != nil {
		log.Printf("failed to clear Redis queue: %v", err)
	} else {
		log.Println("✅ Redis queue cleared")
	}
}

func TestDirectInsert(t *testing.T) {
	rate, rErr := strconv.Atoi(getEnv("INSERT_RATE", "1000"))
	if rErr != nil {
		t.Fatal(rErr)
	}
	total, cErr := strconv.Atoi(getEnv("INSERT_COUNT", "1000"))
	if cErr != nil {
		t.Fatal(cErr)
	}
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := context.Background()
	start := time.Now()
	startFormatted := start.Format("15:04:05")
	t.Logf("Test started at: %s\n", startFormatted)

	for i := 0; i < total; i++ {
		<-ticker.C
		ev, err := RandomEventData()
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		err = ProcessEventDirect(ctx, ev)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
	}

	end := time.Now()
	endFormatted := end.Format("15:04:05")
	duration := end.Sub(start)

	// Format duration as minutes:seconds
	durationFormatted := fmt.Sprintf("%d:%02d", int(duration.Minutes()), int(duration.Seconds())%60)

	actualRPS := float64(total) / duration.Seconds()

	t.Logf("Test started at: %s\n", startFormatted)
	t.Logf("Test ended at: %s\n", endFormatted)
	t.Logf("Test duration: %s\n", durationFormatted)
	t.Logf("Direct Insert: %d events inserted (%.2f RPS)\n", total, actualRPS)

	cleanup()
}

func TestBulkInsert(t *testing.T) {
	rate, rErr := strconv.Atoi(getEnv("INSERT_RATE", "1000"))
	if rErr != nil {
		t.Fatal(rErr)
	}
	total, cErr := strconv.Atoi(getEnv("INSERT_COUNT", "1000"))
	if cErr != nil {
		t.Fatal(cErr)
	}
	interval := time.Second / time.Duration(rate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx := context.Background()
	start := time.Now()
	startFormatted := start.Format("15:04:05")
	t.Logf("Test started at: %s\n", startFormatted)

	for i := 0; i < total; i++ {
		<-ticker.C
		ev, err := RandomEventData()
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
		err = ProcessEventBulk(ctx, ev)
		if err != nil {
			t.Fatalf("failed: %v", err)
		}
	}

	// Wait for remaining events to be processed
	time.Sleep(2 * time.Second)

	duration := time.Since(start)
	actualRPS := float64(total) / duration.Seconds()

	end := time.Now()
	endFormatted := end.Format("15:04:05")

	t.Logf("Test ended at: %s\n", endFormatted)
	t.Logf("Bulk Insert: %d events inserted in %v (%.2f RPS)\n", total, duration, actualRPS)
	t.Logf("Total processed via bulk: %d\n", totalProcessed.Load())

	cleanup()
}

func BenchmarkDirectInsert(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ev, err := RandomEventData()
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
		err = ProcessEventDirect(ctx, ev)
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
	}
}

func BenchmarkBulkInsert(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ev, err := RandomEventData()
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
		err = ProcessEventBulk(ctx, ev)
		if err != nil {
			b.Fatalf("failed: %v", err)
		}
	}
}

func BenchmarkDirectInsertParallel(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev, err := RandomEventData()
			if err != nil {
				b.Fatalf("failed: %v", err)
			}
			err = ProcessEventDirect(ctx, ev)
			if err != nil {
				b.Fatalf("failed: %v", err)
			}
		}
	})
}

func BenchmarkBulkInsertParallel(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ev, err := RandomEventData()
			if err != nil {
				b.Fatalf("failed: %v", err)
			}
			err = ProcessEventBulk(ctx, ev)
			if err != nil {
				b.Fatalf("failed: %v", err)
			}
		}
	})
}

type RedisBulkInserter struct {
	redisClient *redis.Client
	db          *pgxpool.Pool
	queueName   string
	batchSize   int
}

func NewRedisBulkInserter(redisClient *redis.Client, db *pgxpool.Pool, batchSize int) *RedisBulkInserter {
	return &RedisBulkInserter{
		redisClient: redisClient,
		db:          db,
		queueName:   "webhook_events",
		batchSize:   batchSize,
	}
}

func (r *RedisBulkInserter) Save(ctx context.Context, event *eventpb.Event) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = r.redisClient.RPush(ctx, r.queueName, payload).Result()
	if err != nil {
		return fmt.Errorf("failed to push to Redis: %w", err)
	}

	queueLength, err := r.redisClient.LLen(ctx, r.queueName).Result()
	if err != nil {
		return fmt.Errorf("failed to get queue length: %w", err)
	}

	if queueLength >= int64(r.batchSize) {
		go func() {
			err := r.ProcessBatch(context.Background())
			if err != nil {
				log.Printf("Failed to process batch: %v", err)
			}
		}()
	}

	return nil
}

func (r *RedisBulkInserter) ProcessBatch(ctx context.Context) error {
	events, err := r.getBatchFromRedis(ctx)
	if err != nil {
		return fmt.Errorf("failed to get batch from Redis: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	err = r.bulkInsertPostgresSQL(ctx, events)
	if err != nil {
		r.requeueFailedEvents(ctx, events)
		return fmt.Errorf("bulk insert failed: %w", err)
	}

	totalProcessed.Add(int64(len(events)))

	return nil
}

func (r *RedisBulkInserter) requeueFailedEvents(ctx context.Context, events []string) {
	for _, event := range events {
		r.redisClient.LPush(ctx, r.queueName, event)
	}
}

func (r *RedisBulkInserter) getBatchFromRedis(ctx context.Context) ([]string, error) {
	pipe := r.redisClient.TxPipeline()

	eventsCmd := pipe.LRange(ctx, r.queueName, 0, int64(r.batchSize)-1)
	pipe.LTrim(ctx, r.queueName, int64(r.batchSize), -1)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis transaction failed: %w", err)
	}

	return eventsCmd.Val(), nil
}

func (r *RedisBulkInserter) bulkInsertPostgresSQL(ctx context.Context, events []string) error {
	if len(events) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, payload := range events {
		var event eventpb.Event
		err := json.Unmarshal([]byte(payload), &event)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: %w", err)
		}

		jsonPayload, err := convertEventToJSON(&event)

		batch.Queue(`
			INSERT INTO webhook_events 
			(provider, delivery_id, event_type, payload, received_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (provider, delivery_id) DO NOTHING
		`,
			event.Provider,
			event.Id,
			event.EventName,
			jsonPayload,
			time.Now(),
		)
	}

	results := r.db.SendBatch(ctx, batch)
	defer func(results pgx.BatchResults) {
		err := results.Close()
		if err != nil {

		}
	}(results)

	for i := 0; i < batch.Len(); i++ {
		results.Exec()
	}

	return nil
}

func (r *RedisBulkInserter) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond) // Check more frequently for high RPS
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			queueLength, err := r.redisClient.LLen(ctx, r.queueName).Result()
			if err != nil {
				log.Printf("Failed to get queue length: %v", err)
				continue
			}

			if queueLength > 0 {
				err := r.ProcessBatch(ctx)
				if err != nil {
					log.Printf("Failed to process batch: %v", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

/* Event Processing Functions */

func ProcessEventDirect(ctx context.Context, ev *eventpb.Event) error {
	jsonPayload, err := convertEventToJSON(ev)

	_, err = pgPool.Exec(ctx,
		`
			INSERT INTO webhook_events 
			(provider, delivery_id, event_type, payload, received_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (provider, delivery_id) DO NOTHING`,
		ev.Provider,
		ev.Id,
		ev.EventName,
		jsonPayload,
		time.Now(),
	)

	return err
}

func ProcessEventBulk(ctx context.Context, ev *eventpb.Event) error {
	return redisBulkInserter.Save(ctx, ev)
}

/* Event Generation Functions */

func randomEvent() delivery.EventType {
	possibleEvents := []delivery.EventType{
		delivery.EventTypeIssues,
		//service.EventTypeIssueComment,
		//service.EventTypePullRequest,
		//service.EventTypePullRequestReview,
		//service.EventTypePush,
	}
	return possibleEvents[rand.Intn(len(possibleEvents))]
}

func RandomEventData() (*eventpb.Event, error) {
	eventName := randomEvent()
	eventId := randomHex(12)
	repositoryId := rand.Intn(1000000) + 1000
	repositoryName := fmt.Sprintf("%s/%s", randomUserName(), randomRepoName())
	issueId := rand.Intn(1_000_000)
	issueNumber := rand.Intn(1000) + 1
	senderId := rand.Intn(1000000) + 500
	now := time.Now()

	var ev eventpb.Event

	switch eventName {
	case delivery.EventTypeIssues:
		possibleActions := []string{"opened", "closed"}
		randomAction := possibleActions[rand.Intn(len(possibleActions))]
		switch randomAction {
		case "opened":
			ev = eventpb.Event{
				Id:             eventId,
				EventName:      eventpb.EventName_EVENT_NAME_ISSUE_OPENED,
				Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
				Time:           timestamppb.New(now),
				RepositoryId:   uint64(repositoryId),
				RepositoryName: repositoryName,
				Payload: &eventpb.Event_IssueOpenedPayload{
					IssueOpenedPayload: &eventpb.IssueOpenedPayload{
						UserId:      uint64(senderId),
						IssueNumber: int32(issueNumber),
						Title:       randomTitle(),
					},
				},
			}
		case "closed":
			ev = eventpb.Event{
				Id:             eventId,
				EventName:      eventpb.EventName_EVENT_NAME_ISSUE_CLOSED,
				Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
				Time:           timestamppb.New(now),
				RepositoryId:   uint64(repositoryId),
				RepositoryName: repositoryName,
				Payload: &eventpb.Event_IssueClosedPayload{
					IssueClosedPayload: &eventpb.IssueClosedPayload{
						UserId:        uint64(senderId),
						IssueAuthorId: uint64(senderId),
						IssueId:       uint64(issueId),
						IssueNumber:   int32(issueNumber),
						CloseReason:   eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_COMPLETED,
						Labels:        []string{},
						OpenedAt:      timestamppb.New(now.Add(-time.Duration(rand.Intn(72)) * time.Hour)),
						CommentsCount: int32(rand.Intn(10)),
					},
				},
			}
		}
	}

	if ev.Id == "" {
		return nil, fmt.Errorf("invalid event generated")
	}

	return &ev, nil
}

func randomHex(n int) string {
	buf := make([]byte, n/2)
	rand.Read(buf)
	return fmt.Sprintf("%x", buf)
}

func randomAlphaNum(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randomChoice[T any](arr []T) T {
	return arr[rand.Intn(len(arr))]
}

func randomUserName() string {
	return fmt.Sprintf("u-%s", randomAlphaNum(6))
}

func randomRepoName() string {
	return fmt.Sprintf("repo-%s", randomAlphaNum(5))
}

func randomTitle() string {
	words := []string{"Fix", "Add", "Update", "Remove", "Refactor", "Improve", "Document"}
	return fmt.Sprintf("%s %s", randomChoice(words), randomAlphaNum(4))
}

func convertEventToJSON(event *eventpb.Event) (string, error) {
	// Create a structured JSON object that includes both event metadata and payload
	jsonData := map[string]interface{}{
		"id":              event.Id,
		"event_name":      event.EventName.String(),
		"repository_id":   event.RepositoryId,
		"repository_name": event.RepositoryName,
		"provider":        event.Provider.String(),
		"payload":         extractPayloadData(event),
	}

	jsonBytes, err := json.Marshal(jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	return string(jsonBytes), nil
}

func extractPayloadData(event *eventpb.Event) interface{} {
	switch payload := event.Payload.(type) {
	case *eventpb.Event_PrOpenedPayload:
		return map[string]interface{}{
			"type": "pull_request_opened",
			"data": map[string]interface{}{
				"pr_number": payload.PrOpenedPayload.PrNumber,
				"pr_title":  payload.PrOpenedPayload.Title,
			},
		}

	case *eventpb.Event_PrClosedPayload:
		return map[string]interface{}{
			"type": "pull_request_closed",
			"data": map[string]interface{}{
				"pr_number": payload.PrClosedPayload.PrNumber,
			},
		}

	case *eventpb.Event_PrReviewPayload:
		return map[string]interface{}{
			"type": "pull_request_review",
			"data": map[string]interface{}{
				"pr_number": payload.PrReviewPayload.PrNumber,
			},
		}

	case *eventpb.Event_IssueOpenedPayload:
		return map[string]interface{}{
			"type": "issue_opened",
			"data": map[string]interface{}{
				"issue_number": payload.IssueOpenedPayload.IssueNumber,
			},
		}

	case *eventpb.Event_IssueClosedPayload:
		return map[string]interface{}{
			"type": "issue_closed",
			"data": map[string]interface{}{
				"issue_number": payload.IssueClosedPayload.IssueNumber,
			},
		}

	case *eventpb.Event_IssueCommentedPayload:
		return map[string]interface{}{
			"type": "issue_commented",
			"data": map[string]interface{}{
				"issue_number": payload.IssueCommentedPayload.IssueNumber,
			},
		}

	case *eventpb.Event_PushPayload:
		return map[string]interface{}{
			"type": "push",
			"data": map[string]interface{}{
				"branch":       payload.PushPayload.BranchName,
				"commit_count": payload.PushPayload.CommitsCount,
			},
		}

	default:
		return map[string]interface{}{
			"type": "unknown",
			"data": nil,
		}
	}
}
