package cachemanager

import (
	"context"
	"errors"
	"testing"
	"time"
	"unsafe"

	"github.com/gocasters/rankr/adapter/redis"
	redis_client "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient mocks the redis.Client methods we need
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis_client.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis_client.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis_client.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis_client.StringCmd)
}

// Helper functions to create mock commands with proper behavior
func createStatusCmd(err error) *redis_client.StatusCmd {
	cmd := redis_client.NewStatusCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal("OK")
	}
	return cmd
}

func createStringCmd(result string, err error) *redis_client.StringCmd {
	cmd := redis_client.NewStringCmd(context.Background())
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(result)
	}
	return cmd
}

// createMockAdapter creates a mock redis.Adapter by using unsafe pointer manipulation
func createMockAdapter(mockClient *MockRedisClient) *redis.Adapter {
	// Create a real adapter first (this will be empty/nil)
	var adapter redis.Adapter

	// Use unsafe to set the internal client field
	// This is a bit hacky but allows us to inject our mock
	adapterPtr := unsafe.Pointer(&adapter)
	clientFieldPtr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(adapterPtr)))
	*clientFieldPtr = unsafe.Pointer(mockClient)

	return &adapter
}

func TestCacheManager_Set(t *testing.T) {
	ctx := context.Background()

	// Test case: Set succeeds
	t.Run("Set_Success", func(t *testing.T) {
		mockClient := &MockRedisClient{}
		mockClient.On("Set", ctx, "key1", "value1", time.Minute).Return(createStatusCmd(nil))

		mockAdapter := createMockAdapter(mockClient)
		cacheManager := NewCacheManager(mockAdapter)

		err := cacheManager.Set(ctx, "key1", "value1", time.Minute)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	// Test case: Set returns error
	t.Run("Set_Error", func(t *testing.T) {
		mockClient := &MockRedisClient{}
		expectedErr := errors.New("redis set error")
		mockClient.On("Set", ctx, "key2", "value2", time.Minute).Return(createStatusCmd(expectedErr))

		mockAdapter := createMockAdapter(mockClient)
		cacheManager := NewCacheManager(mockAdapter)

		err := cacheManager.Set(ctx, "key2", "value2", time.Minute)
		assert.Error(t, err)
		assert.Equal(t, "redis set error", err.Error())
		mockClient.AssertExpectations(t)
	})
}

func TestCacheManager_Get(t *testing.T) {
	ctx := context.Background()

	// Test case: Get succeeds
	t.Run("Get_Success", func(t *testing.T) {
		mockClient := &MockRedisClient{}
		mockClient.On("Get", ctx, "key1").Return(createStringCmd("value1", nil))

		mockAdapter := createMockAdapter(mockClient)
		cacheManager := NewCacheManager(mockAdapter)

		result, err := cacheManager.Get(ctx, "key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", result)
		mockClient.AssertExpectations(t)
	})

	// Test case: Get returns error
	t.Run("Get_Error", func(t *testing.T) {
		mockClient := &MockRedisClient{}
		expectedErr := errors.New("redis get error")
		mockClient.On("Get", ctx, "key2").Return(createStringCmd("", expectedErr))

		mockAdapter := createMockAdapter(mockClient)
		cacheManager := NewCacheManager(mockAdapter)

		result, err := cacheManager.Get(ctx, "key2")
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Equal(t, "redis get error", err.Error())
		mockClient.AssertExpectations(t)
	})
}
