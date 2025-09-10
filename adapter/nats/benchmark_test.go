package nats

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func BenchmarkConfigValidate_Performance(b *testing.B) {
	config := validConfig

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfigSetDefaults_Performance(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		config := Config{}
		config.SetDefaults()
	}
}

func BenchmarkFormatValidationErrors_Performance(b *testing.B) {
	errors := map[string]error{
		"url":              fmt.Errorf("invalid URL"),
		"client_id":        fmt.Errorf("invalid client ID"),
		"timeout":          fmt.Errorf("invalid timeout"),
		"max_inflight":     fmt.Errorf("invalid max inflight"),
		"reconnect_wait":   fmt.Errorf("invalid reconnect wait"),
		"ping_interval":    fmt.Errorf("invalid ping interval"),
		"max_pings_out":    fmt.Errorf("invalid max pings out"),
		"ack_wait_timeout": fmt.Errorf("invalid ack wait timeout"),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = FormatValidationErrors(errors)
	}
}

func BenchmarkNewAdapter_ConfigValidation(b *testing.B) {
	ctx := context.Background()
	logger := watermill.NopLogger{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {

		config := validConfig
		config.URL = "nats://invalid-host:4222"

		_, _ = New(ctx, config, logger)
	}
}

func BenchmarkMessageCreation(b *testing.B) {
	payload := []byte(`{"event_id": "test-123", "timestamp": "2023-01-01T00:00:00Z", "data": "benchmark payload data"}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		msg := message.NewMessage(watermill.NewUUID(), payload)
		msg.Metadata.Set("event_type", "benchmark")
		msg.Metadata.Set("source", "test")
		msg.Metadata.Set("timestamp", time.Now().Format(time.RFC3339))
	}
}

func BenchmarkUUIDGeneration(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = watermill.NewUUID()
	}
}

func BenchmarkConfigValidate_Concurrent(b *testing.B) {
	config := validConfig
	numWorkers := 10

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < b.N/numWorkers; i++ {
				_ = config.Validate()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkConfigValidate_InvalidConfig(b *testing.B) {

	configs := []Config{
		{URL: "", ClientID: "test"},
		{URL: "nats://localhost:4222", ClientID: ""},
		{URL: "nats://localhost:4222", ClientID: "test", MaxInflight: 0},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		config := configs[i%len(configs)]
		_ = config.Validate()
	}
}

func BenchmarkAdapterMethods_NilComponents(b *testing.B) {
	adapter := &Adapter{
		config:     validConfig,
		publisher:  nil,
		subscriber: nil,
		conn:       nil,
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.Run("IsConnected", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.IsConnected()
		}
	})

	b.Run("Status", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.Status()
		}
	})

	b.Run("Config", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.Config()
		}
	})

	b.Run("Publisher", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.Publisher()
		}
	})

	b.Run("Subscriber", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = adapter.Subscriber()
		}
	})
}

func BenchmarkErrorScenarios(b *testing.B) {
	adapter := &Adapter{
		config:     validConfig,
		publisher:  nil,
		subscriber: nil,
		conn:       nil,
	}

	ctx := context.Background()

	b.Run("Publish_NilPublisher", func(b *testing.B) {
		msg := message.NewMessage("test", []byte("payload"))
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = adapter.Publish("topic", msg)
		}
	})

	b.Run("Subscribe_NilSubscriber", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = adapter.Subscribe(ctx, "topic")
		}
	})

	b.Run("Flush_NilConnection", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = adapter.Flush()
		}
	})

	b.Run("FlushTimeout_NilConnection", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = adapter.FlushTimeout(5 * time.Second)
		}
	})
}

func BenchmarkConfigScenarios(b *testing.B) {
	scenarios := map[string]Config{
		"CoreNATS": {
			URL:            "nats://localhost:4222",
			ClientID:       "benchmark-core",
			UseJetStream:   false,
			AckWaitTimeout: 30 * time.Second,
			MaxInflight:    1024,
			ConnectTimeout: 5 * time.Second,
			PingInterval:   2 * time.Minute,
			MaxPingsOut:    2,
		},
		"JetStream": {
			URL:            "nats://localhost:4222",
			ClientID:       "benchmark-js",
			UseJetStream:   true,
			DurableName:    "benchmark",
			AckWaitTimeout: 30 * time.Second,
			MaxInflight:    1024,
			ConnectTimeout: 5 * time.Second,
			PingInterval:   2 * time.Minute,
			MaxPingsOut:    2,
		},
		"HighThroughput": {
			URL:            "nats://localhost:4222",
			ClientID:       "benchmark-ht",
			UseJetStream:   true,
			DurableName:    "benchmark-ht",
			AckWaitTimeout: 5 * time.Second,
			MaxInflight:    10000,
			ConnectTimeout: 10 * time.Second,
			PingInterval:   30 * time.Second,
			MaxPingsOut:    10,
		},
	}

	for name, config := range scenarios {
		b.Run(name+"_Validate", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = config.Validate()
			}
		})

		b.Run(name+"_SetDefaults", func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				testConfig := config
				testConfig.SetDefaults()
			}
		})
	}
}

func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("ConfigCreation", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = &Config{
				URL:            "nats://localhost:4222",
				ClientID:       "benchmark",
				DurableName:    "test",
				QueueGroup:     "test-group",
				AckWaitTimeout: 30 * time.Second,
				MaxInflight:    1024,
				ConnectTimeout: 5 * time.Second,
				ReconnectWait:  2 * time.Second,
				MaxReconnects:  -1,
				PingInterval:   2 * time.Minute,
				MaxPingsOut:    2,
				AllowReconnect: true,
				UseJetStream:   true,
			}
		}
	})

	b.Run("AdapterCreation", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = &Adapter{
				config:     validConfig,
				publisher:  nil,
				subscriber: nil,
				conn:       nil,
				logger:     watermill.NopLogger{},
			}
		}
	})

	b.Run("ErrorMapCreation", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			errs := map[string]error{
				"url":            fmt.Errorf("error 1"),
				"client_id":      fmt.Errorf("error 2"),
				"timeout":        fmt.Errorf("error 3"),
				"max_inflight":   fmt.Errorf("error 4"),
				"reconnect_wait": fmt.Errorf("error 5"),
			}
			_ = FormatValidationErrors(errs)
		}
	})
}

func BenchmarkConcurrentOperations(b *testing.B) {
	adapter := &Adapter{
		config:     validConfig,
		publisher:  nil,
		subscriber: nil,
		conn:       nil,
		logger:     watermill.NopLogger{},
	}

	b.Run("ConcurrentConfigAccess", func(b *testing.B) {
		numWorkers := 100
		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N/numWorkers; i++ {
					_ = adapter.Config()
					_ = adapter.IsConnected()
					_ = adapter.Status()
					_ = adapter.Publisher()
					_ = adapter.Subscriber()
				}
			}()
		}
		wg.Wait()
	})

	b.Run("ConcurrentValidation", func(b *testing.B) {
		configs := []Config{validConfig, validConfig, validConfig}
		numWorkers := 50
		b.ResetTimer()
		b.ReportAllocs()

		var wg sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				config := configs[workerID%len(configs)]
				for i := 0; i < b.N/numWorkers; i++ {
					_ = config.Validate()
				}
			}(w)
		}
		wg.Wait()
	})
}

func BenchmarkScaling(b *testing.B) {
	workerCounts := []int{1, 10, 50, 100, 500}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			config := validConfig
			b.ResetTimer()
			b.ReportAllocs()

			var wg sync.WaitGroup
			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N/workers; i++ {
						errors := config.Validate()
						if len(errors) > 0 {
							_ = FormatValidationErrors(errors)
						}
					}
				}()
			}
			wg.Wait()
		})
	}
}

func BenchmarkComparison(b *testing.B) {

	b.Run("ValidationWithMap", func(b *testing.B) {
		config := validConfig
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = config.Validate()
		}
	})

	b.Run("ValidationWithSlice", func(b *testing.B) {
		config := validConfig
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var errors []error
			if config.URL == "" {
				errors = append(errors, fmt.Errorf("URL is empty"))
			}
			if config.ClientID == "" {
				errors = append(errors, fmt.Errorf("ClientID is empty"))
			}

			_ = errors
		}
	})
}

func BenchmarkRealWorldScenarios(b *testing.B) {
	b.Run("TypicalApplicationStartup", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {

			config := Config{
				URL:      "nats://localhost:4222",
				ClientID: fmt.Sprintf("app-%d", i),
			}
			config.SetDefaults()
			errors := config.Validate()
			if len(errors) > 0 {
				_ = FormatValidationErrors(errors)
			}
		}
	})

	b.Run("ConfigurationReload", func(b *testing.B) {
		configs := make([]Config, 100)
		for i := range configs {
			configs[i] = Config{
				URL:      "nats://localhost:4222",
				ClientID: fmt.Sprintf("reload-test-%d", i),
			}
			configs[i].SetDefaults()
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			config := configs[i%len(configs)]
			_ = config.Validate()
		}
	})
}
