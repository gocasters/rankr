package leaderboardscoring_test

import (
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValidator_ValidateEvent_Success(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	event := &leaderboardscoring.EventRequest{
		ID:             uuid.New().String(),
		UserID:         "12345",
		EventName:      leaderboardscoring.PullRequestOpened.String(),
		RepositoryID:   1001,
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
		Payload: leaderboardscoring.PullRequestOpenedPayload{
			UserID:       12345,
			PrID:         1,
			Title:        "Test",
			BranchName:   "feature/test",
			TargetBranch: "main",
		},
	}

	err := validator.ValidateEvent(event)
	assert.NoError(t, err)
}

func TestValidator_ValidateEvent_MissingID(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	event := &leaderboardscoring.EventRequest{
		ID:             "",
		EventName:      leaderboardscoring.PullRequestOpened.String(),
		RepositoryID:   1001,
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
	}

	err := validator.ValidateEvent(event)
	assert.Error(t, err)
}

func TestValidator_ValidateEvent_InvalidUUID(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	event := &leaderboardscoring.EventRequest{
		ID:             "not-a-uuid",
		EventName:      leaderboardscoring.PullRequestOpened.String(),
		RepositoryID:   1001,
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
	}

	err := validator.ValidateEvent(event)
	assert.Error(t, err)
}

func TestValidator_ValidateEvent_InvalidEventName(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	event := &leaderboardscoring.EventRequest{
		ID:             uuid.New().String(),
		EventName:      "invalid_event",
		RepositoryID:   1001,
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
	}

	err := validator.ValidateEvent(event)
	assert.Error(t, err)
}

func TestValidator_ValidateEvent_MissingRepository(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	event := &leaderboardscoring.EventRequest{
		ID:             uuid.New().String(),
		EventName:      leaderboardscoring.PullRequestOpened.String(),
		RepositoryID:   0, // Missing
		RepositoryName: "test-repo",
		Timestamp:      time.Now().UTC(),
	}

	err := validator.ValidateEvent(event)
	assert.Error(t, err)
}

func TestValidator_ValidateGetLeaderboard_Success(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Monthly.String(),
		PageSize:  10,
		Offset:    1,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.NoError(t, err)
}

func TestValidator_ValidateGetLeaderboard_WithProjectID(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	projectID := "my-project"
	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Weekly.String(),
		ProjectID: &projectID,
		PageSize:  50,
		Offset:    10,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.NoError(t, err)
}

func TestValidator_ValidateGetLeaderboard_InvalidTimeframe(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: "invalid",
		PageSize:  50,
		Offset:    10,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.Error(t, err)
}

func TestValidator_ValidateGetLeaderboard_MissingPageSize(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Monthly.String(),
		PageSize:  0,
		Offset:    10,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.Error(t, err)
}

func TestValidator_ValidateGetLeaderboard_InvalidPageSize(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Monthly.String(),
		PageSize:  1_001,
		Offset:    10,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.Error(t, err)
}

func TestValidator_ValidateGetLeaderboard_InvalidOffset(t *testing.T) {
	validator := leaderboardscoring.NewValidator()

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Monthly.String(),
		PageSize:  20,
		Offset:    -1,
	}

	err := validator.ValidateGetLeaderboard(req)
	assert.Error(t, err)
}
