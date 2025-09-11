package repository

import (
	"context"
	"fmt"
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

// WebhookRepositoryTestSuite provides a test suite with database setup/teardown
type WebhookRepositoryTestSuite struct {
	suite.Suite
	db   *pgxpool.Pool
	repo WebhookRepository
	ctx  context.Context
}

func TestWebhookRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(WebhookRepositoryTestSuite))
}

func (suite *WebhookRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Get database URL from environment variable
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://webhook_admin:password123@localhost:5436/webhook_db?sslmode=disable"
		suite.T().Logf("DATABASE_URL not set, using default: %s", dbURL)
	}

	var err error
	suite.db, err = pgxpool.New(suite.ctx, dbURL)
	require.NoError(suite.T(), err, "Failed to connect to test database")

	// Test connection
	err = suite.db.Ping(suite.ctx)
	require.NoError(suite.T(), err, "Failed to ping test database")

	suite.repo = NewWebhookRepository(suite.db)

	// Create table for tests
	suite.createTable()
}

func (suite *WebhookRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.dropTable()
		suite.db.Close()
	}
}

func (suite *WebhookRepositoryTestSuite) SetupTest() {
	// Clean table before each test
	_, err := suite.db.Exec(suite.ctx, "TRUNCATE TABLE webhook_events RESTART IDENTITY")
	require.NoError(suite.T(), err)
}

func (suite *WebhookRepositoryTestSuite) createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS webhook_events (
		id BIGSERIAL PRIMARY KEY,
		provider smallint NOT NULL,
		delivery_id TEXT NOT NULL,
		event_type smallint NOT NULL,
		payload BYTEA NOT NULL,
		received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		CONSTRAINT webhook_events_provider_delivery_id_unique UNIQUE (provider, delivery_id)
	)`
	_, err := suite.db.Exec(suite.ctx, query)
	require.NoError(suite.T(), err, "Failed to create test table")
}

func (suite *WebhookRepositoryTestSuite) dropTable() {
	_, err := suite.db.Exec(suite.ctx, "DROP TABLE IF EXISTS webhook_events")
	require.NoError(suite.T(), err, "Failed to drop test table")
}

// Helper functions to create different types of test events
func (suite *WebhookRepositoryTestSuite) createPullRequestOpenedEvent(deliveryID string, userID uint64) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       userID,
				PrId:         98765,
				PrNumber:     42,
				Title:        "Add awesome feature",
				BranchName:   "feature/awesome",
				TargetBranch: "main",
				Labels:       []string{"enhancement", "priority-high"},
				Assignees:    []uint64{userID, 999},
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createPullRequestClosedEvent(deliveryID string, userID uint64, merged bool) *eventpb.Event {
	closeReason := eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED
	if !merged {
		closeReason = eventpb.PrCloseReason_PR_CLOSE_REASON_CLOSED_WITHOUT_MERGE
	}

	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrClosedPayload{
			PrClosedPayload: &eventpb.PullRequestClosedPayload{
				UserId:       userID,
				MergerUserId: userID,
				PrId:         98765,
				PrNumber:     42,
				CloseReason:  closeReason,
				Merged:       &merged,
				Additions:    150,
				Deletions:    25,
				FilesChanged: 8,
				CommitsCount: 5,
				Labels:       []string{"enhancement", "priority-high"},
				TargetBranch: "main",
				Assignees:    []uint64{userID},
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createPullRequestReviewEvent(deliveryID string, reviewerID, authorID uint64, state eventpb.ReviewState) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrReviewPayload{
			PrReviewPayload: &eventpb.PullRequestReviewSubmittedPayload{
				ReviewerUserId: reviewerID,
				PrAuthorUserId: authorID,
				PrId:           98765,
				PrNumber:       42,
				State:          state,
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createIssueOpenedEvent(deliveryID string, userID uint64) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueOpenedPayload{
			IssueOpenedPayload: &eventpb.IssueOpenedPayload{
				UserId:      userID,
				IssueId:     54321,
				IssueNumber: 123,
				Title:       "Bug found in login system",
				Labels:      []string{"bug", "priority-critical"},
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createIssueClosedEvent(deliveryID string, userID, authorID uint64) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_CLOSED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueClosedPayload{
			IssueClosedPayload: &eventpb.IssueClosedPayload{
				UserId:          userID,
				IssueAuthorId:   authorID,
				IssueId:         54321,
				IssueNumber:     123,
				CloseReason:     eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_COMPLETED,
				Labels:          []string{"bug", "fixed"},
				OpenedAt:        timestamppb.New(time.Now().Add(-24 * time.Hour)),
				CommentsCount:   5,
				ClosingPrNumber: 42,
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createIssueCommentedEvent(deliveryID string, userID, authorID uint64) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_IssueCommentedPayload{
			IssueCommentedPayload: &eventpb.IssueCommentedPayload{
				UserId:        userID,
				IssueAuthorId: authorID,
				IssueId:       54321,
				IssueNumber:   123,
				CommentLength: 250,
				ContainsCode:  true,
			},
		},
	}
}

func (suite *WebhookRepositoryTestSuite) createPushEvent(deliveryID string, userID uint64) *eventpb.Event {
	return &eventpb.Event{
		Id:             deliveryID,
		EventName:      eventpb.EventName_EVENT_NAME_PUSHED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PushPayload{
			PushPayload: &eventpb.PushPayload{
				UserId:       userID,
				BranchName:   "main",
				CommitsCount: 3,
				Commits: []*eventpb.CommitInfo{
					{
						AuthorName: "John Doe",
						CommitId:   "abc123def456",
						Message:    "Fix critical bug",
						Additions:  50,
						Deletions:  10,
						Modified:   3,
					},
					{
						AuthorName: "John Doe",
						CommitId:   "def456ghi789",
						Message:    "Update documentation",
						Additions:  25,
						Deletions:  5,
						Modified:   2,
					},
					{
						AuthorName: "John Doe",
						CommitId:   "ghi789jkl012",
						Message:    "Add tests",
						Additions:  100,
						Deletions:  0,
						Modified:   5,
					},
				},
			},
		},
	}
}

// Test Save method with different event types
func (suite *WebhookRepositoryTestSuite) TestSave_PullRequestOpened() {
	event := suite.createPullRequestOpenedEvent("pr-opened-1", 1001)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	// Verify the event was saved
	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_PullRequestClosed() {
	event := suite.createPullRequestClosedEvent("pr-closed-1", 1001, true)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_PullRequestReview() {
	event := suite.createPullRequestReviewEvent("pr-review-1", 2001, 1001, eventpb.ReviewState_REVIEW_STATE_APPROVED)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_IssueOpened() {
	event := suite.createIssueOpenedEvent("issue-opened-1", 1001)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_IssueClosed() {
	event := suite.createIssueClosedEvent("issue-closed-1", 1001, 2001)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_IssueCommented() {
	event := suite.createIssueCommentedEvent("issue-comment-1", 1001, 2001)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_Push() {
	event := suite.createPushEvent("push-1", 1001)

	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

func (suite *WebhookRepositoryTestSuite) TestSave_DuplicateIgnored() {
	event := suite.createPullRequestOpenedEvent("duplicate-delivery", 1001)

	// Save first time
	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	// Save duplicate - should return error but not fail
	err = suite.repo.Save(suite.ctx, event)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate webhook event")

	// Verify only one event exists
	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), count)
}

// Test FindByDeliveryID method
func (suite *WebhookRepositoryTestSuite) TestFindByDeliveryID_Success() {
	originalEvent := suite.createIssueOpenedEvent("delivery-test", 1001)

	err := suite.repo.Save(suite.ctx, originalEvent)
	require.NoError(suite.T(), err)

	foundEvent, err := suite.repo.FindByDeliveryID(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), "delivery-test")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), foundEvent)
	assert.Equal(suite.T(), originalEvent.Provider, foundEvent.Provider)
	assert.Equal(suite.T(), originalEvent.Id, foundEvent.Id)

	// Verify payload was correctly deserialized
	assert.NotNil(suite.T(), foundEvent.GetIssueOpenedPayload())
	issuePayload := foundEvent.GetIssueOpenedPayload()
	assert.Equal(suite.T(), uint64(1001), issuePayload.UserId)
	assert.Equal(suite.T(), uint64(54321), issuePayload.IssueId)
	assert.Equal(suite.T(), int32(123), issuePayload.IssueNumber)
}

func (suite *WebhookRepositoryTestSuite) TestFindByDeliveryID_NotFound() {
	foundEvent, err := suite.repo.FindByDeliveryID(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), "non-existent")
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), foundEvent)
}

// Test FindEvents with filters
func (suite *WebhookRepositoryTestSuite) TestFindEvents_WithProviderFilter() {
	// Create events with different providers (though we only have GitHub in enum)
	event1 := suite.createPullRequestOpenedEvent("delivery-1", 1001)
	event2 := suite.createIssueOpenedEvent("delivery-2", 1002)

	err := suite.repo.Save(suite.ctx, event1)
	require.NoError(suite.T(), err)
	err = suite.repo.Save(suite.ctx, event2)
	require.NoError(suite.T(), err)

	// Filter by GitHub provider
	provider := int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB)
	filter := EventFilter{Provider: &provider}

	events, err := suite.repo.FindEvents(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 2) // Both should be GitHub

	for _, event := range events {
		assert.Equal(suite.T(), eventpb.EventProvider_EVENT_PROVIDER_GITHUB, event.Provider)
	}
}

func (suite *WebhookRepositoryTestSuite) TestFindEvents_WithEventTypeFilter() {
	event1 := suite.createPullRequestOpenedEvent("delivery-1", 1001)
	event2 := suite.createIssueOpenedEvent("delivery-2", 1001)

	err := suite.repo.Save(suite.ctx, event1)
	require.NoError(suite.T(), err)
	err = suite.repo.Save(suite.ctx, event2)
	require.NoError(suite.T(), err)

	eventType := int32(eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED)
	filter := EventFilter{EventType: &eventType}

	events, err := suite.repo.FindEvents(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 1)
	assert.Equal(suite.T(), eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED, events[0].EventName)
}

func (suite *WebhookRepositoryTestSuite) TestFindEvents_WithTimeRangeFilter() {
	// Save an event
	event := suite.createPushEvent("time-test", 1001)
	err := suite.repo.Save(suite.ctx, event)
	require.NoError(suite.T(), err)

	// Filter with time range that includes the event
	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	filter := EventFilter{
		StartTime: &start,
		EndTime:   &end,
	}

	events, err := suite.repo.FindEvents(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 1)

	// Filter with time range that excludes the event
	pastStart := now.Add(-2 * time.Hour)
	pastEnd := now.Add(-1 * time.Hour)

	filter = EventFilter{
		StartTime: &pastStart,
		EndTime:   &pastEnd,
	}

	events, err = suite.repo.FindEvents(suite.ctx, filter)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 0)
}

// Test convenience methods
func (suite *WebhookRepositoryTestSuite) TestGetEventsByProvider() {
	event1 := suite.createPullRequestOpenedEvent("delivery-1", 1001)
	event2 := suite.createIssueOpenedEvent("delivery-2", 1002)

	err := suite.repo.Save(suite.ctx, event1)
	require.NoError(suite.T(), err)
	err = suite.repo.Save(suite.ctx, event2)
	require.NoError(suite.T(), err)

	events, err := suite.repo.GetEventsByProvider(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 2) // Both are GitHub events

	for _, event := range events {
		assert.Equal(suite.T(), eventpb.EventProvider_EVENT_PROVIDER_GITHUB, event.Provider)
	}
}

func (suite *WebhookRepositoryTestSuite) TestGetEventsByType() {
	event1 := suite.createPullRequestOpenedEvent("delivery-1", 1001)
	event2 := suite.createIssueOpenedEvent("delivery-2", 1001)

	err := suite.repo.Save(suite.ctx, event1)
	require.NoError(suite.T(), err)
	err = suite.repo.Save(suite.ctx, event2)
	require.NoError(suite.T(), err)

	events, err := suite.repo.GetEventsByType(suite.ctx, int32(eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED), 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 1)
	assert.Equal(suite.T(), eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED, events[0].EventName)
}

func (suite *WebhookRepositoryTestSuite) TestGetRecentEvents() {
	// Create events with delays to ensure different timestamps
	for i := 0; i < 3; i++ {
		var event *eventpb.Event
		switch i {
		case 0:
			event = suite.createPullRequestOpenedEvent(fmt.Sprintf("recent-%d", i), uint64(1001+i))
		case 1:
			event = suite.createIssueOpenedEvent(fmt.Sprintf("recent-%d", i), uint64(1001+i))
		case 2:
			event = suite.createPushEvent(fmt.Sprintf("recent-%d", i), uint64(1001+i))
		}

		err := suite.repo.Save(suite.ctx, event)
		require.NoError(suite.T(), err)

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}

	events, err := suite.repo.GetRecentEvents(suite.ctx, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), events, 2)

	// Events should be ordered by received_at DESC (most recent first)
	assert.Equal(suite.T(), "recent-2", events[0].Id)
	assert.Equal(suite.T(), "recent-1", events[1].Id)
}

// Test utility methods
func (suite *WebhookRepositoryTestSuite) TestEventExists() {
	// Should not exist initially
	exists, err := suite.repo.EventExists(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), "exists-test")
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)

	// Save event
	event := suite.createPullRequestOpenedEvent("exists-test", 1001)
	err = suite.repo.Save(suite.ctx, event)
	require.NoError(suite.T(), err)

	// Should exist now
	exists, err = suite.repo.EventExists(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), "exists-test")
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// Test edge cases and error scenarios
func (suite *WebhookRepositoryTestSuite) TestSave_AllEventTypes() {
	events := []*eventpb.Event{
		suite.createPullRequestOpenedEvent("all-types-1", 1001),
		suite.createPullRequestClosedEvent("all-types-2", 1002, true),
		suite.createPullRequestReviewEvent("all-types-3", 2001, 1001, eventpb.ReviewState_REVIEW_STATE_APPROVED),
		suite.createIssueOpenedEvent("all-types-4", 1003),
		suite.createIssueClosedEvent("all-types-5", 1004, 1003),
		suite.createIssueCommentedEvent("all-types-6", 1005, 1003),
		suite.createPushEvent("all-types-7", 1006),
	}

	for _, event := range events {
		err := suite.repo.Save(suite.ctx, event)
		assert.NoError(suite.T(), err, "Failed to save event type: %s", event.EventName)
	}

	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(7), count)

	// Verify we can retrieve and deserialize all event types
	savedEvents, err := suite.repo.GetRecentEvents(suite.ctx, 10)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), savedEvents, 7)

	// Check that all payloads are properly deserialized
	eventTypesSeen := make(map[eventpb.EventName]bool)
	for _, savedEvent := range savedEvents {
		eventTypesSeen[savedEvent.EventName] = true

		switch savedEvent.EventName {
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED:
			assert.NotNil(suite.T(), savedEvent.GetPrOpenedPayload())
			assert.Equal(suite.T(), "Add awesome feature", savedEvent.GetPrOpenedPayload().Title)
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED:
			assert.NotNil(suite.T(), savedEvent.GetPrClosedPayload())
			assert.Equal(suite.T(), int32(42), savedEvent.GetPrClosedPayload().PrNumber)
		case eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED:
			assert.NotNil(suite.T(), savedEvent.GetPrReviewPayload())
			assert.Equal(suite.T(), eventpb.ReviewState_REVIEW_STATE_APPROVED, savedEvent.GetPrReviewPayload().State)
		case eventpb.EventName_EVENT_NAME_ISSUE_OPENED:
			assert.NotNil(suite.T(), savedEvent.GetIssueOpenedPayload())
			assert.Equal(suite.T(), "Bug found in login system", savedEvent.GetIssueOpenedPayload().Title)
		case eventpb.EventName_EVENT_NAME_ISSUE_CLOSED:
			assert.NotNil(suite.T(), savedEvent.GetIssueClosedPayload())
			assert.Equal(suite.T(), eventpb.IssueCloseReason_ISSUE_CLOSE_REASON_COMPLETED, savedEvent.GetIssueClosedPayload().CloseReason)
		case eventpb.EventName_EVENT_NAME_ISSUE_COMMENTED:
			assert.NotNil(suite.T(), savedEvent.GetIssueCommentedPayload())
			assert.Equal(suite.T(), int32(250), savedEvent.GetIssueCommentedPayload().CommentLength)
		case eventpb.EventName_EVENT_NAME_PUSHED:
			assert.NotNil(suite.T(), savedEvent.GetPushPayload())
			assert.Equal(suite.T(), "main", savedEvent.GetPushPayload().BranchName)
			assert.Len(suite.T(), savedEvent.GetPushPayload().Commits, 3)
		}
	}

	// Verify we saw all expected event types
	assert.Equal(suite.T(), 7, len(eventTypesSeen))
}

// Performance and stress tests
func (suite *WebhookRepositoryTestSuite) TestSavePerformance_BatchInserts() {
	start := time.Now()

	// Test saving 100 different event types
	for i := 0; i < 100; i++ {
		var event *eventpb.Event
		switch i % 7 {
		case 0:
			event = suite.createPullRequestOpenedEvent(fmt.Sprintf("perf-pr-opened-%d", i), uint64(1001+i))
		case 1:
			event = suite.createPullRequestClosedEvent(fmt.Sprintf("perf-pr-closed-%d", i), uint64(1001+i), i%2 == 0)
		case 2:
			event = suite.createPullRequestReviewEvent(fmt.Sprintf("perf-pr-review-%d", i), uint64(2001+i), uint64(1001+i), eventpb.ReviewState_REVIEW_STATE_APPROVED)
		case 3:
			event = suite.createIssueOpenedEvent(fmt.Sprintf("perf-issue-opened-%d", i), uint64(1001+i))
		case 4:
			event = suite.createIssueClosedEvent(fmt.Sprintf("perf-issue-closed-%d", i), uint64(1001+i), uint64(2001+i))
		case 5:
			event = suite.createIssueCommentedEvent(fmt.Sprintf("perf-issue-comment-%d", i), uint64(1001+i), uint64(2001+i))
		case 6:
			event = suite.createPushEvent(fmt.Sprintf("perf-push-%d", i), uint64(1001+i))
		}

		err := suite.repo.Save(suite.ctx, event)
		require.NoError(suite.T(), err)
	}

	duration := time.Since(start)
	suite.T().Logf("Saved 100 mixed events in %v (avg: %v per event)",
		duration, duration/100)

	// Verify all events were saved
	count, err := suite.repo.CountEvents(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(100), count)

	// Basic sanity check - shouldn't take more than 10 seconds for 100 events
	assert.Less(suite.T(), duration, 10*time.Second,
		"Saving 100 events took too long: %v", duration)
}

// Test repository method compatibility with your existing repository interface
func (suite *WebhookRepositoryTestSuite) TestRepositoryInterfaceCompatibility() {
	// This test ensures our repository methods work with the expected data structure
	// based on your original Save method signature

	event := &eventpb.Event{
		Id:             "interface-test",
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.Now(),
		RepositoryId:   12345,
		RepositoryName: "testorg/testrepo",
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       1001,
				PrId:         98765,
				PrNumber:     42,
				Title:        "Interface test PR",
				BranchName:   "feature/test",
				TargetBranch: "main",
				Labels:       []string{"test"},
				Assignees:    []uint64{1001},
			},
		},
	}

	// Test that Save works with our event structure
	err := suite.repo.Save(suite.ctx, event)
	assert.NoError(suite.T(), err)

	// Test that we can retrieve and it matches expected structure
	savedEvent, err := suite.repo.FindByDeliveryID(suite.ctx, int32(eventpb.EventProvider_EVENT_PROVIDER_GITHUB), "interface-test")
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), savedEvent)

	// Verify all fields are properly saved and retrieved
	assert.Equal(suite.T(), event.Id, savedEvent.Id)
	assert.Equal(suite.T(), event.EventName, savedEvent.EventName)
	assert.Equal(suite.T(), event.RepositoryId, savedEvent.RepositoryId)
	assert.Equal(suite.T(), event.RepositoryName, savedEvent.RepositoryName)
	assert.Equal(suite.T(), event.Provider, savedEvent.Provider)

	// Verify payload is correctly deserialized
	assert.NotNil(suite.T(), savedEvent.GetPrOpenedPayload())
	originalPayload := event.GetPrOpenedPayload()
	savedPayload := savedEvent.GetPrOpenedPayload()

	assert.Equal(suite.T(), originalPayload.UserId, savedPayload.UserId)
	assert.Equal(suite.T(), originalPayload.PrId, savedPayload.PrId)
	assert.Equal(suite.T(), originalPayload.PrNumber, savedPayload.PrNumber)
	assert.Equal(suite.T(), originalPayload.Title, savedPayload.Title)
	assert.Equal(suite.T(), originalPayload.BranchName, savedPayload.BranchName)
	assert.Equal(suite.T(), originalPayload.TargetBranch, savedPayload.TargetBranch)
	assert.Equal(suite.T(), originalPayload.Labels, savedPayload.Labels)
	assert.Equal(suite.T(), originalPayload.Assignees, savedPayload.Assignees)
}
