package cachemanager

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

// MockClient mocks redis.Client
type MockClient struct {
    SetFunc func(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd
    GetFunc func(ctx context.Context, key string) *MockStringCmd
}

func (m *MockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
    return m.SetFunc(ctx, key, value, expiration)
}

func (m *MockClient) Get(ctx context.Context, key string) *MockStringCmd {
    return m.GetFunc(ctx, key)
}

// MockStatusCmd mocks redis.StatusCmd
type MockStatusCmd struct {
    err error
}

func (m *MockStatusCmd) Err() error {
    return m.err
}

// MockStringCmd mocks redis.StringCmd
type MockStringCmd struct {
    result string
    err    error
}

func (m *MockStringCmd) Result() (string, error) {
    return m.result, m.err
}

// MockAdapter mocks redis.Adapter
type MockAdapter struct {
    client *MockClient
}

func (m *MockAdapter) Client() *MockClient {
    return m.client
}

func TestCacheManager_Set(t *testing.T) {
    ctx := context.Background()

    // Test case: Set succeeds
    mockClient := &MockClient{
        SetFunc: func(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
            return &MockStatusCmd{err: nil}
        },
    }
    mockAdapter := &MockAdapter{client: mockClient}
    cacheManager := NewCacheManager(mockAdapter)

    err := cacheManager.Set(ctx, "key1", "value1", time.Minute)
    assert.NoError(t, err)

    // Test case: Set returns error
    mockClient.SetFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) *MockStatusCmd {
        return &MockStatusCmd{err: errors.New("redis set error")}
    }

    err = cacheManager.Set(ctx, "key2", "value2", time.Minute)
    assert.Error(t, err)
    assert.Equal(t, "redis set error", err.Error())
}

func TestCacheManager_Get(t *testing.T) {
    ctx := context.Background()

    // Test case: Get succeeds
    mockClient := &MockClient{
        GetFunc: func(ctx context.Context, key string) *MockStringCmd {
            return &MockStringCmd{result: "value1", err: nil}
        },
    }
    mockAdapter := &MockAdapter{client: mockClient}
    cacheManager := NewCacheManager(mockAdapter)

    val, err := cacheManager.Get(ctx, "key1")
    assert.NoError(t, err)
    assert.Equal(t, "value1", val)

    // Test case: Get returns error
    mockClient.GetFunc = func(ctx context.Context, key string) *MockStringCmd {
        return &MockStringCmd{result: "", err: errors.New("redis get error")}
    }

    val, err = cacheManager.Get(ctx, "key2")
    assert.Error(t, err)
    assert.Equal(t, "", val)
}
