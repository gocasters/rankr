package userprofile

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *mockLeaderboardStatRPC) GetContributorStat(ctx context.Context, userID int64) (ContributorStat, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ContributorStat), args.Error(1)
}

// --- Tests ---

func TestContributorProfile_Success(t *testing.T) {
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

	svc := NewService(mockInfo, mockTask, mockStat, validator)

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

func TestContributorProfile_Error(t *testing.T) {
	ctx := context.Background()
	userID := int64(123)

	mockInfo := new(mockContributorRPC)
	mockInfo.On("GetProfileInfo", mock.Anything, userID).Return(ContributorInfo{}, errors.New("failed"))

	mockTask := new(mockTaskRPC)
	mockTask.On("GetTasks", mock.Anything, userID).Return([]Task{}, nil)

	mockStat := new(mockLeaderboardStatRPC)
	mockStat.On("GetContributorStat", mock.Anything, userID).Return(ContributorStat{}, nil)

	validator := NewValidator(nil)

	svc := NewService(mockInfo, mockTask, mockStat, validator)

	resp, err := svc.ContributorProfile(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockInfo.AssertExpectations(t)
}
