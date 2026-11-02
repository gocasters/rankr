package userprofile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mock types ---

type mockContributorRPC struct {
	mock.Mock
}

func (m *mockContributorRPC) GetProfileInfo(ctx context.Context, userID int64) (ContributorInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ContributorInfo), args.Error(1)
}

type mockTaskRPC struct {
	mock.Mock
}

func (m *mockTaskRPC) GetTasks(ctx context.Context, userID int64) ([]Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]Task), args.Error(1)
}

type mockLeaderboardStatRPC struct {
	mock.Mock
}

type blockingContributorRPC struct {
	started chan struct{}
	result  chan error
}

func (m blockingContributorRPC) GetProfileInfo(ctx context.Context, userID int64) (ContributorInfo, error) {
	close(m.started)
	<-ctx.Done()
	err := ctx.Err()
	m.result <- err
	return ContributorInfo{}, err
}

func (m *mockLeaderboardStatRPC) GetContributorStat(ctx context.Context, userID int64) (ContributorStat, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ContributorStat), args.Error(1)
}

// --- Tests ---

func TestContributorProfile_Success(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))

	ctx := context.Background()
	userID := int64(123)

	// mock data
	expectedInfo := ContributorInfo{GitHubUsername: "John Doe"}
	expectedTasks := []Task{{ID: 1}, {ID: 2}}
	expectedStat := ContributorStat{GlobalRank: 5}

	// set up mocks
	mockInfo := new(mockContributorRPC)
	mockInfo.On("GetProfileInfo", mock.Anything, userID).Return(expectedInfo, nil)

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return(expectedTasks, nil)

	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(expectedStat, nil)

	// use real validator (does nothing)
	validator := NewValidator(nil)

	svc := NewService(mockInfo, mockTask, mockStat, validator, Config{UpstreamTimeout: 100 * time.Millisecond})

	resp, err := svc.ContributorProfile(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedInfo, resp.Profile.ContributorInfo)
	assert.Equal(t, expectedTasks, resp.Profile.Tasks)
	assert.Equal(t, expectedStat, resp.Profile.ContributorStat)

	mockInfo.AssertExpectations(t)
	mockTask.AssertExpectations(t)
	mockStat.AssertExpectations(t)
}

func TestContributorProfile_ContributorFailureReturnsPartialProfile(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))

	ctx := context.Background()
	userID := int64(123)
	expectedTasks := []Task{{ID: 1, Title: "task"}}
	expectedStat := ContributorStat{GlobalRank: 5}

	mockInfo := new(mockContributorRPC)
	mockInfo.On("GetProfileInfo", mock.Anything, userID).Return(ContributorInfo{}, errors.New("failed"))

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return(expectedTasks, nil)

	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(expectedStat, nil)

	validator := NewValidator(nil)

	svc := NewService(mockInfo, mockTask, mockStat, validator, Config{UpstreamTimeout: 100 * time.Millisecond})

	resp, err := svc.ContributorProfile(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, ContributorInfo{}, resp.Profile.ContributorInfo)
	assert.Equal(t, expectedTasks, resp.Profile.Tasks)
	assert.Equal(t, expectedStat, resp.Profile.ContributorStat)
	mockInfo.AssertExpectations(t)
	mockTask.AssertExpectations(t)
	mockStat.AssertExpectations(t)
}

func TestContributorProfile_LeaderboardStatFailureReturnsPartialProfile(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))

	ctx := context.Background()
	userID := int64(123)
	expectedInfo := ContributorInfo{GitHubUsername: "John Doe"}
	expectedTasks := []Task{{ID: 1, Title: "task"}}

	mockInfo := new(mockContributorRPC)
	mockInfo.On("GetProfileInfo", mock.Anything, userID).Return(expectedInfo, nil)

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return(expectedTasks, nil)

	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(ContributorStat{}, errors.New("failed"))

	svc := NewService(mockInfo, mockTask, mockStat, NewValidator(nil), Config{UpstreamTimeout: 100 * time.Millisecond})

	resp, err := svc.ContributorProfile(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, expectedInfo, resp.Profile.ContributorInfo)
	assert.Equal(t, expectedTasks, resp.Profile.Tasks)
	assert.Equal(t, ContributorStat{}, resp.Profile.ContributorStat)
	mockInfo.AssertExpectations(t)
	mockTask.AssertExpectations(t)
	mockStat.AssertExpectations(t)
}

func TestContributorProfile_TaskFailureReturnsPartialProfile(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))

	ctx := context.Background()
	userID := int64(123)
	expectedInfo := ContributorInfo{GitHubUsername: "John Doe"}
	expectedStat := ContributorStat{GlobalRank: 5}

	mockInfo := new(mockContributorRPC)
	mockInfo.On("GetProfileInfo", mock.Anything, userID).Return(expectedInfo, nil)

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return([]Task(nil), errors.New("failed"))

	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(expectedStat, nil)

	svc := NewService(mockInfo, mockTask, mockStat, NewValidator(nil), Config{UpstreamTimeout: 100 * time.Millisecond})

	resp, err := svc.ContributorProfile(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, expectedInfo, resp.Profile.ContributorInfo)
	assert.Equal(t, []Task{}, resp.Profile.Tasks)
	assert.Equal(t, expectedStat, resp.Profile.ContributorStat)
	mockInfo.AssertExpectations(t)
	mockTask.AssertExpectations(t)
	mockStat.AssertExpectations(t)
}

func TestContributorProfile_UpstreamTimeoutReturnsPartialProfile(t *testing.T) {
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))

	userID := int64(123)
	expectedTasks := []Task{{ID: 1, Title: "task"}}
	expectedStat := ContributorStat{GlobalRank: 5}
	started := make(chan struct{})
	result := make(chan error, 1)
	blockingInfo := blockingContributorRPC{started: started, result: result}

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return(expectedTasks, nil)
	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(expectedStat, nil)

	svc := NewService(blockingInfo, mockTask, mockStat, NewValidator(nil), Config{UpstreamTimeout: 50 * time.Millisecond})

	resp, err := svc.ContributorProfile(context.Background(), userID)

	require.NoError(t, err)
	require.NotNil(t, resp)
	select {
	case <-started:
	default:
		t.Fatal("blocking contributor upstream was not called")
	}
	require.ErrorIs(t, <-result, context.DeadlineExceeded)
	assert.Equal(t, ContributorInfo{}, resp.Profile.ContributorInfo)
	assert.Equal(t, expectedTasks, resp.Profile.Tasks)
	assert.Equal(t, expectedStat, resp.Profile.ContributorStat)
	mockTask.AssertExpectations(t)
	mockStat.AssertExpectations(t)
}
