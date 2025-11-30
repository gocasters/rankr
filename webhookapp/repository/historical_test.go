package repository

import (
	"context"
	"os"
	"testing"
	"time"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HistoricalEventTestSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	repo WebhookRepository
	ctx  context.Context
}

func TestHistoricalEventTestSuite(t *testing.T) {
	suite.Run(t, new(HistoricalEventTestSuite))
}

func (suite *HistoricalEventTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://webhook_admin:password123@localhost:5436/webhook_db?sslmode=disable"
		suite.T().Logf("DATABASE_URL not set, using default: %s", dbURL)
	}

	var err error
	suite.db, err = pgxpool.New(suite.ctx, dbURL)
	require.NoError(suite.T(), err, "Failed to connect to test database")

	err = suite.db.Ping(suite.ctx)

	require.NoError(suite.T(), err, "Failed to ping test database")

	suite.repo = NewWebhookRepository(suite.db)

	suite.createTableWithMigration()
}

func (suite *HistoricalEventTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.dropTable()
		suite.db.Close()
	}
}

func (suite *HistoricalEventTestSuite) SetupTest() {
	_, err := suite.db.Exec(suite.ctx, "TRUNCATE TABLE webhook_events RESTART IDENTITY")
	require.NoError(suite.T(), err)
}

func (suite *HistoricalEventTestSuite) createTableWithMigration() {
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS webhook_events (
		id BIGSERIAL PRIMARY KEY,
		provider smallint NOT NULL,
		delivery_id TEXT,
		event_type smallint NOT NULL,
		payload BYTEA NOT NULL,
		received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		source VARCHAR(20) DEFAULT 'webhook',
		resource_type VARCHAR(20),
		resource_id BIGINT,
		event_key TEXT
	)`
	_, err := suite.db.Exec(suite.ctx, createTableQuery)
	require.NoError(suite.T(), err, "Failed to create test table")

	suite.db.Exec(suite.ctx, "DROP INDEX IF EXISTS webhook_events_webhook_unique_idx")
	suite.db.Exec(suite.ctx, "DROP INDEX IF EXISTS webhook_events_historical_unique_idx")
	suite.db.Exec(suite.ctx, "DROP INDEX IF EXISTS webhook_events_global_unique_idx")
	suite.db.Exec(suite.ctx, "DROP INDEX IF EXISTS webhook_events_delivery_unique_idx")
	suite.db.Exec(suite.ctx, "DROP INDEX IF EXISTS webhook_events_event_key_unique_idx")
	suite.db.Exec(suite.ctx, "ALTER TABLE webhook_events DROP CONSTRAINT IF EXISTS webhook_events_provider_delivery_id_unique")

	eventKeyIndexQuery := `
	CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_event_key_unique_idx
	ON webhook_events(event_key)
	WHERE event_key IS NOT NULL`
	_, err = suite.db.Exec(suite.ctx, eventKeyIndexQuery)
	require.NoError(suite.T(), err, "Failed to create event_key unique index")

	deliveryIndexQuery := `
	CREATE UNIQUE INDEX IF NOT EXISTS webhook_events_delivery_unique_idx
	ON webhook_events(provider, delivery_id)
	WHERE delivery_id IS NOT NULL`
	_, err = suite.db.Exec(suite.ctx, deliveryIndexQuery)
	require.NoError(suite.T(), err, "Failed to create delivery unique index")
}

func (suite *HistoricalEventTestSuite) dropTable() {
	_, err := suite.db.Exec(suite.ctx, "DROP TABLE IF EXISTS webhook_events")
	require.NoError(suite.T(), err, "Failed to drop test table")
}

func (suite *HistoricalEventTestSuite) createHistoricalPROpenedEvent(prNumber int32) *eventpb.Event {
	return &eventpb.Event{
		Id:             "not-used-for-historical",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   999,
		RepositoryName: "gocasters/rankr",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       1001,
				PrId:         12345,
				PrNumber:     prNumber,
				Title:        "Historical PR",
				BranchName:   "feature/test",
				TargetBranch: "main",
				Labels:       []string{"enhancement"},
				Assignees:    []uint64{1001},
			},
		},
	}
}

func (suite *HistoricalEventTestSuite) createHistoricalPRClosedEvent(prNumber int32) *eventpb.Event {
	merged := true
	return &eventpb.Event{
		Id:             "not-used-for-historical",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   999,
		RepositoryName: "gocasters/rankr",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrClosedPayload{
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId:       1001,
				MergerUserId: 1001,
				PrId:         12345,
				PrNumber:     prNumber,
				CloseReason:  eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED,
				Merged:       &merged,
				Additions:    100,
				Deletions:    20,
				FilesChanged: 5,
				CommitsCount: 3,
				Labels:       []string{"enhancement"},
				TargetBranch: "main",
				Assignees:    []uint64{1001},
			},
		},
	}
}

func (suite *HistoricalEventTestSuite) createHistoricalReviewEvent(prNumber int32, reviewID int64) *eventpb.Event {
	return &eventpb.Event{
		Id:             "not-used-for-historical",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		Time:           timestamppb.Now(),
		RepositoryId:   999,
		RepositoryName: "gocasters/rankr",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrReviewPayload{
			PrReviewPayload: &eventpb.PullRequestReviewSubmittedPayload{
				ReviewerUserId: 2001,
				PrAuthorUserId: 1001,
				PrId:           12345,
				PrNumber:       prNumber,
				State:          eventpb.ReviewState_REVIEW_STATE_APPROVED,
			},
		},
	}
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_FirstTime_Success() {
	event := suite.createHistoricalPROpenedEvent(42)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, event, "pull_request", 42)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)

	var source string
	var resourceType string
	var resourceID int64
	err = suite.db.QueryRow(suite.ctx,
		"SELECT source, resource_type, resource_id FROM webhook_events WHERE event_type=$1",
		eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
	).Scan(&source, &resourceType, &resourceID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "historical", source)
	assert.Equal(suite.T(), "pull_request", resourceType)
	assert.Equal(suite.T(), int64(42), resourceID)
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_Duplicate_Rejected() {
	event := suite.createHistoricalPROpenedEvent(42)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, event, "pull_request", 42)
	require.NoError(suite.T(), err, "First save should succeed")

	err = suite.repo.SaveHistoricalEvent(suite.ctx, event, "pull_request", 42)
	assert.Error(suite.T(), err, "Duplicate save should fail")
	assert.Equal(suite.T(), ErrDuplicateEvent, err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count, "Should still have only 1 event")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_DifferentEventTypes_SamePR_Allowed() {
	openedEvent := suite.createHistoricalPROpenedEvent(42)
	closedEvent := suite.createHistoricalPRClosedEvent(42)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, openedEvent, "pull_request", 42)
	require.NoError(suite.T(), err, "Save opened event should succeed")

	err = suite.repo.SaveHistoricalEvent(suite.ctx, closedEvent, "pull_request", 42)
	assert.NoError(suite.T(), err, "Save closed event should succeed (different event_type)")

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), count, "Should have 2 events (opened + closed)")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_DifferentPRs_SameEventType_Allowed() {
	event1 := suite.createHistoricalPROpenedEvent(42)
	event2 := suite.createHistoricalPROpenedEvent(43)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, event1, "pull_request", 42)
	require.NoError(suite.T(), err)

	err = suite.repo.SaveHistoricalEvent(suite.ctx, event2, "pull_request", 43)
	assert.NoError(suite.T(), err, "Different PR numbers should be allowed")

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), count)
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_ReviewEvents_Idempotency() {
	review1 := suite.createHistoricalReviewEvent(42, 7890)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, review1, "pull_request_review", 7890)
	require.NoError(suite.T(), err)

	err = suite.repo.SaveHistoricalEvent(suite.ctx, review1, "pull_request_review", 7890)
	assert.Error(suite.T(), err, "Duplicate review should be rejected")
	assert.Equal(suite.T(), ErrDuplicateEvent, err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *HistoricalEventTestSuite) TestWebhookAndHistoricalEvents_ShouldConflict() {
	webhookEvent := &eventpb.Event{
		Id:             "webhook-delivery-123",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   999,
		RepositoryName: "gocasters/rankr",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       1001,
				PrId:         12345,
				PrNumber:     42,
				Title:        "Webhook PR",
				BranchName:   "feature/webhook",
				TargetBranch: "main",
			},
		},
	}

	err := suite.repo.Save(suite.ctx, webhookEvent)
	require.NoError(suite.T(), err, "Webhook event should save successfully")

	historicalEvent := suite.createHistoricalPROpenedEvent(42)
	err = suite.repo.SaveHistoricalEvent(suite.ctx, historicalEvent, "pull_request", 42)
	assert.Error(suite.T(), err, "Historical event with same PR should conflict with webhook event")
	assert.Equal(suite.T(), ErrDuplicateEvent, err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count, "Should have only webhook event")
}

func (suite *HistoricalEventTestSuite) TestHistoricalThenWebhook_ShouldConflict() {
	historicalEvent := suite.createHistoricalPROpenedEvent(42)
	err := suite.repo.SaveHistoricalEvent(suite.ctx, historicalEvent, "pull_request", 42)
	require.NoError(suite.T(), err, "Historical event should save successfully")

	webhookEvent := &eventpb.Event{
		Id:             "webhook-delivery-123",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   999,
		RepositoryName: "gocasters/rankr",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       1001,
				PrId:         12345,
				PrNumber:     42,
				Title:        "Webhook PR",
				BranchName:   "feature/webhook",
				TargetBranch: "main",
			},
		},
	}

	err = suite.repo.Save(suite.ctx, webhookEvent)
	assert.Error(suite.T(), err, "Webhook event with same PR should conflict with historical event")
	assert.Equal(suite.T(), ErrDuplicateEvent, err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count, "Should have only historical event")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_DeliveryIDIsNull() {
	event := suite.createHistoricalPROpenedEvent(42)

	err := suite.repo.SaveHistoricalEvent(suite.ctx, event, "pull_request", 42)
	require.NoError(suite.T(), err)

	var deliveryID *string
	err = suite.db.QueryRow(suite.ctx,
		"SELECT delivery_id FROM webhook_events WHERE source='historical' AND resource_id=42",
	).Scan(&deliveryID)

	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), deliveryID, "delivery_id should be NULL for historical events")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_CompleteWorkflow() {
	prNumber := int32(100)

	openedEvent := suite.createHistoricalPROpenedEvent(prNumber)
	err := suite.repo.SaveHistoricalEvent(suite.ctx, openedEvent, "pull_request", int64(prNumber))
	require.NoError(suite.T(), err, "Save PR opened event")

	review1 := suite.createHistoricalReviewEvent(prNumber, 7890)
	err = suite.repo.SaveHistoricalEvent(suite.ctx, review1, "pull_request_review", 7890)
	require.NoError(suite.T(), err, "Save first review")

	review2 := suite.createHistoricalReviewEvent(prNumber, 7891)
	err = suite.repo.SaveHistoricalEvent(suite.ctx, review2, "pull_request_review", 7891)
	require.NoError(suite.T(), err, "Save second review")

	closedEvent := suite.createHistoricalPRClosedEvent(prNumber)
	err = suite.repo.SaveHistoricalEvent(suite.ctx, closedEvent, "pull_request", int64(prNumber))
	require.NoError(suite.T(), err, "Save PR closed event")

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(4), count, "Should have 4 events: opened + 2 reviews + closed")

	err = suite.repo.SaveHistoricalEvent(suite.ctx, openedEvent, "pull_request", int64(prNumber))
	assert.Error(suite.T(), err, "Duplicate opened event should be rejected")

	err = suite.repo.SaveHistoricalEvent(suite.ctx, review1, "pull_request_review", 7890)
	assert.Error(suite.T(), err, "Duplicate review should be rejected")

	count, err = suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(4), count, "Count should remain 4 after duplicate attempts")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_ConcurrentSaves_NoDuplicates() {
	event := suite.createHistoricalPROpenedEvent(50)

	done := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func() {
			err := suite.repo.SaveHistoricalEvent(suite.ctx, event, "pull_request", 50)
			done <- err
		}()
	}

	var successCount, duplicateCount int
	for i := 0; i < 5; i++ {
		err := <-done
		if err == nil {
			successCount++
		} else if err == ErrDuplicateEvent {
			duplicateCount++
		} else {
			suite.T().Errorf("Unexpected error: %v", err)
		}
	}

	assert.Equal(suite.T(), 1, successCount, "Exactly 1 save should succeed")
	assert.Equal(suite.T(), 4, duplicateCount, "Other 4 should return duplicate error")

	time.Sleep(100 * time.Millisecond)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count, "Should have exactly 1 event despite concurrent attempts")
}

func (suite *HistoricalEventTestSuite) TestSaveHistoricalEvent_MultipleReposNoConflict() {
	event1 := suite.createHistoricalPROpenedEvent(42)
	event1.RepositoryId = 111
	event1.RepositoryName = "org1/repo1"

	event2 := suite.createHistoricalPROpenedEvent(42)
	event2.RepositoryId = 222
	event2.RepositoryName = "org2/repo2"
	event2.Provider = eventpb.EventProvider_EVENT_PROVIDER_GITHUB

	err := suite.repo.SaveHistoricalEvent(suite.ctx, event1, "pull_request", 42)
	require.NoError(suite.T(), err)

	err = suite.repo.SaveHistoricalEvent(suite.ctx, event2, "pull_request", 42)
	assert.Error(suite.T(), err, "Same provider+resource_type+resource_id+event_type should conflict even with different repos")

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}
