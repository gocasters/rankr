package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

var validConfig = Config{
	Host:     "localhost",
	Port:     6379,
	Password: "",
	DB:       0,
}

func TestConfigValidate_Valid(t *testing.T) {
	vErrors := validConfig.Validate()
	assert.Empty(t, vErrors)
}

func TestConfigValidate_Invalid(t *testing.T) {
	testCases := []struct {
		name           string
		config         Config
		expectedErrors []string
	}{
		{
			name:           "Empty Host",
			config:         Config{Host: "", Port: 6379, DB: 0},
			expectedErrors: []string{"host"},
		},
		{
			name:           "Invalid Port (negative)",
			config:         Config{Host: "localhost", Port: -1, DB: 0},
			expectedErrors: []string{"port"},
		},
		{
			name:           "Invalid Port (too high)",
			config:         Config{Host: "localhost", Port: 65536, DB: 0},
			expectedErrors: []string{"port"},
		},
		{
			name:           "Invalid DB (negative)",
			config:         Config{Host: "localhost", Port: 6379, DB: -1},
			expectedErrors: []string{"db"},
		},
		{
			name:           "Invalid DB (too high)",
			config:         Config{Host: "localhost", Port: 6379, DB: 16},
			expectedErrors: []string{"db"},
		},
		{
			name:           "Multiple errors",
			config:         Config{Host: "", Port: -1, DB: -1},
			expectedErrors: []string{"host", "port", "db"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vErrors := tc.config.Validate()

			assert.Equal(t, len(tc.expectedErrors), len(vErrors))

			for _, expectedField := range tc.expectedErrors {
				assert.Contains(t, vErrors, expectedField)
				assert.Error(t, vErrors[expectedField])
			}
		})
	}
}

func TestFormatValidationErrors(t *testing.T) {
	testCases := []struct {
		name     string
		errors   map[string]error
		expected string
	}{
		{
			name:     "No errors",
			errors:   map[string]error{},
			expected: "",
		},
		{
			name: "Single error",
			errors: map[string]error{
				"host": errors.New("redis host is empty"),
			},
			expected: "validation errors: host: redis host is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatValidationErrors(tc.errors)
			if tc.name == "No errors" {
				assert.Equal(t, tc.expected, result)
			} else {
				assert.Contains(t, result, "validation errors:")
				assert.Contains(t, result, "host: redis host is empty")
			}
		})
	}
}

func TestNew_WithMock(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		defer func() {
			if closeErr := client.Close(); closeErr != nil {
				t.Errorf("Failed to close mock client: %v", closeErr)
			}
		}()

		mock.ExpectPing().SetVal("PONG")

		adapter := &Adapter{client: client}

		err := adapter.client.Ping(context.Background()).Err()
		assert.NoError(t, err)
		assert.NotNil(t, adapter)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Connection Failure", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		defer func() {
			if closeErr := client.Close(); closeErr != nil {
				t.Errorf("Failed to close mock client: %v", closeErr)
			}
		}()

		mock.ExpectPing().SetErr(errors.New("connection refused"))

		err := client.Ping(context.Background()).Err()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestNew_InvalidConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      Config
		expectedErr string
	}{
		{
			name:        "Empty Host",
			config:      Config{Host: "", Port: 6379, DB: 0},
			expectedErr: "redis host is empty",
		},
		{
			name:        "Invalid Port (negative)",
			config:      Config{Host: "localhost", Port: -1, DB: 0},
			expectedErr: "invalid redis port: -1",
		},
		{
			name:        "Invalid Port (too high)",
			config:      Config{Host: "localhost", Port: 65536, DB: 0},
			expectedErr: "invalid redis port: 65536",
		},
		{
			name:        "Invalid DB (negative)",
			config:      Config{Host: "localhost", Port: 6379, DB: -1},
			expectedErr: "invalid redis DB: -1",
		},
		{
			name:        "Invalid DB (too high)",
			config:      Config{Host: "localhost", Port: 6379, DB: 16},
			expectedErr: "invalid redis DB: 16",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter, err := New(context.Background(), tc.config)

			assert.Error(t, err)
			assert.Nil(t, adapter)
			assert.Contains(t, err.Error(), "invalid redis configuration")
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

func TestAdapter_Client(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			t.Errorf("Failed to close mock client: %v", closeErr)
		}
	}()

	adapter := &Adapter{client: client}

	assert.Equal(t, client, adapter.Client())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAdapter_Close(t *testing.T) {
	t.Run("Normal Close", func(t *testing.T) {
		client, mock := redismock.NewClientMock()
		adapter := &Adapter{client: client}

		err := adapter.Close()
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Nil Adapter", func(t *testing.T) {
		var adapter *Adapter
		err := adapter.Close()
		assert.NoError(t, err)
	})

	t.Run("Nil Client", func(t *testing.T) {
		adapter := &Adapter{client: nil}
		err := adapter.Close()
		assert.NoError(t, err)
	})
}
