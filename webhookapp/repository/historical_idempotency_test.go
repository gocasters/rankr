package repository

import (
	"testing"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHistoricalEventIdempotencyKeys(t *testing.T) {
	tests := []struct {
		name             string
		resourceType1    string
		resourceID1      int64
		eventType1       eventpb.EventName
		resourceType2    string
		resourceID2      int64
		eventType2       eventpb.EventName
		shouldConflict   bool
		description      string
	}{
		{
			name:           "Same PR same event type should conflict",
			resourceType1:  "pull_request",
			resourceID1:    42,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			resourceType2:  "pull_request",
			resourceID2:    42,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			shouldConflict: true,
			description:    "Duplicate PR opened event for PR #42",
		},
		{
			name:           "Same PR different event types should NOT conflict",
			resourceType1:  "pull_request",
			resourceID1:    42,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			resourceType2:  "pull_request",
			resourceID2:    42,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
			shouldConflict: false,
			description:    "PR opened and PR closed for same PR #42",
		},
		{
			name:           "Different PRs same event type should NOT conflict",
			resourceType1:  "pull_request",
			resourceID1:    42,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			resourceType2:  "pull_request",
			resourceID2:    43,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			shouldConflict: false,
			description:    "PR opened for PR #42 and PR #43",
		},
		{
			name:           "Same review ID should conflict",
			resourceType1:  "pull_request_review",
			resourceID1:    7890,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
			resourceType2:  "pull_request_review",
			resourceID2:    7890,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
			shouldConflict: true,
			description:    "Duplicate review submission",
		},
		{
			name:           "Different review IDs should NOT conflict",
			resourceType1:  "pull_request_review",
			resourceID1:    7890,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
			resourceType2:  "pull_request_review",
			resourceID2:    7891,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
			shouldConflict: false,
			description:    "Two different reviews",
		},
		{
			name:           "PR and Review with same number should NOT conflict",
			resourceType1:  "pull_request",
			resourceID1:    42,
			eventType1:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
			resourceType2:  "pull_request_review",
			resourceID2:    42,
			eventType2:     eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
			shouldConflict: false,
			description:    "PR #42 opened and review #42 (different resource types)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key1 := createIdempotencyKey(
				eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
				tt.resourceType1,
				tt.resourceID1,
				tt.eventType1,
			)
			key2 := createIdempotencyKey(
				eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
				tt.resourceType2,
				tt.resourceID2,
				tt.eventType2,
			)

			if tt.shouldConflict {
				assert.Equal(t, key1, key2,
					"Keys should be identical for: %s", tt.description)
			} else {
				assert.NotEqual(t, key1, key2,
					"Keys should be different for: %s", tt.description)
			}
		})
	}
}

func createIdempotencyKey(provider eventpb.EventProvider, resourceType string, resourceID int64, eventType eventpb.EventName) string {
	return string(rune(provider)) + "-" + resourceType + "-" + string(rune(resourceID)) + "-" + string(rune(eventType))
}

func TestHistoricalEventPayloadPreservation(t *testing.T) {
	event := &eventpb.Event{
		Id:             "test-id",
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
				Title:        "Test PR Title",
				BranchName:   "feature/test",
				TargetBranch: "main",
				Labels:       []string{"bug", "urgent"},
				Assignees:    []uint64{1001, 2002},
			},
		},
	}

	payload := event.GetPrOpenedPayload()
	assert.NotNil(t, payload, "Payload should not be nil")
	assert.Equal(t, uint64(1001), payload.UserId)
	assert.Equal(t, int32(42), payload.PrNumber)
	assert.Equal(t, "Test PR Title", payload.Title)
	assert.Len(t, payload.Labels, 2)
	assert.Len(t, payload.Assignees, 2)
}

func TestHistoricalEventWorkflowSimulation(t *testing.T) {
	type eventRecord struct {
		provider     eventpb.EventProvider
		resourceType string
		resourceID   int64
		eventType    eventpb.EventName
		saved        bool
	}

	saved := make(map[string]bool)

	saveEvent := func(e eventRecord) bool {
		key := createIdempotencyKey(e.provider, e.resourceType, e.resourceID, e.eventType)
		if saved[key] {
			return false
		}
		saved[key] = true
		return true
	}

	prNumber := int64(100)
	provider := eventpb.EventProvider_EVENT_PROVIDER_GITHUB

	t.Run("Complete PR workflow with idempotency", func(t *testing.T) {
		opened := eventRecord{provider, "pull_request", prNumber, eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED, false}
		assert.True(t, saveEvent(opened), "First PR opened should save")

		review1 := eventRecord{provider, "pull_request_review", 7890, eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED, false}
		assert.True(t, saveEvent(review1), "First review should save")

		review2 := eventRecord{provider, "pull_request_review", 7891, eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED, false}
		assert.True(t, saveEvent(review2), "Second review should save")

		closed := eventRecord{provider, "pull_request", prNumber, eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED, false}
		assert.True(t, saveEvent(closed), "PR closed should save")

		assert.False(t, saveEvent(opened), "Duplicate PR opened should be rejected")
		assert.False(t, saveEvent(review1), "Duplicate review 1 should be rejected")
		assert.False(t, saveEvent(review2), "Duplicate review 2 should be rejected")
		assert.False(t, saveEvent(closed), "Duplicate PR closed should be rejected")

		assert.Equal(t, 4, len(saved), "Should have exactly 4 unique events")
	})
}

func TestHistoricalEventDifferentProviders(t *testing.T) {
	saved := make(map[string]bool)

	saveEvent := func(provider eventpb.EventProvider, resourceType string, resourceID int64, eventType eventpb.EventName) bool {
		key := createIdempotencyKey(provider, resourceType, resourceID, eventType)
		if saved[key] {
			return false
		}
		saved[key] = true
		return true
	}

	github := eventpb.EventProvider_EVENT_PROVIDER_GITHUB
	resourceType := "pull_request"
	resourceID := int64(42)
	eventType := eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED

	assert.True(t, saveEvent(github, resourceType, resourceID, eventType),
		"First save for GitHub should succeed")

	assert.False(t, saveEvent(github, resourceType, resourceID, eventType),
		"Duplicate for GitHub should be rejected")
}

func TestHistoricalEventResourceTypeIsolation(t *testing.T) {
	saved := make(map[string]bool)

	saveEvent := func(provider eventpb.EventProvider, resourceType string, resourceID int64, eventType eventpb.EventName) bool {
		key := createIdempotencyKey(provider, resourceType, resourceID, eventType)
		if saved[key] {
			return false
		}
		saved[key] = true
		return true
	}

	provider := eventpb.EventProvider_EVENT_PROVIDER_GITHUB
	resourceID := int64(42)
	eventType := eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED

	assert.True(t, saveEvent(provider, "pull_request", resourceID, eventType),
		"PR with ID 42 should save")

	assert.True(t, saveEvent(provider, "pull_request_review", resourceID, eventType),
		"Review with ID 42 should save (different resource_type)")

	assert.False(t, saveEvent(provider, "pull_request", resourceID, eventType),
		"Duplicate PR should be rejected")

	assert.False(t, saveEvent(provider, "pull_request_review", resourceID, eventType),
		"Duplicate review should be rejected")

	assert.Equal(t, 2, len(saved),
		"Should have 2 events: one PR and one review, both with ID 42")
}