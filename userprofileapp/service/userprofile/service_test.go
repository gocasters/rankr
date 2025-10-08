package userprofile

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type mockRPCRepository struct {
	mock.Mock
}

func (m *mockRPCRepository) GetProfileInfo(ctx context.Context, userID int64) (ContributorInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ContributorInfo), args.Error(1)
}

func (m *mockRPCRepository) GetTasks(ctx context.Context, userID int64) ([]Task, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]Task), args.Error(1)
}

func (m *mockRPCRepository) GetContributorStat(ctx context.Context, userID int64) (ContributorStat, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(ContributorStat), args.Error(1)
}

type mockValidator struct{}

func (m mockValidator) Validate(_ context.Context, _ interface{}) error {
	return nil
}

func TestGetUserProfile_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRPCRepository)
	mockValidator := NewValidator(mockRepo)

	svc := NewService(mockRepo, mockValidator)

	contributor := ContributorInfo{
		ID:             1,
		GitHubID:       12345,
		GitHubUsername: "golang-dev",
		DisplayName:    strPtr("Gopher"),
		CreatedAt:      time.Now(),
	}
	tasks := []Task{
		{ID: 101, Title: "Fix bug", Description: "Fix login bug", State: "open"},
	}
	stats := ContributorStat{
		ContributorID: 1,
		GlobalRank:    5,
		TotalScore:    420.5,
	}

	mockRepo.On("GetProfileInfo", ctx, int64(1)).Return(contributor, nil)
	mockRepo.On("GetTasks", ctx, int64(1)).Return(tasks, nil)
	mockRepo.On("GetContributorStat", ctx, int64(1)).Return(stats, nil)

	resp, err := svc.ContributorProfile(ctx, 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Profile.ContributorInfo.ID)
	assert.Equal(t, "golang-dev", resp.Profile.ContributorInfo.GitHubUsername)
	assert.Len(t, resp.Profile.Tasks, 1)
	assert.Equal(t, 420.5, resp.Profile.ContributorStat.TotalScore)

	mockRepo.AssertExpectations(t)
}

func TestGetUserProfile_Failure_GetTasks(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRPCRepository)
	mockValidator := NewValidator(mockRepo)

	svc := NewService(mockRepo, mockValidator)

	contributor := ContributorInfo{ID: 1, GitHubUsername: "golang-dev"}

	mockRepo.On("GetProfileInfo", ctx, int64(1)).Return(contributor, nil)
	mockRepo.On("GetTasks", ctx, int64(1)).Return([]Task{}, errors.New("task service unavailable"))

	resp, err := svc.ContributorProfile(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}

func strPtr(s string) *string {
	return &s
}
