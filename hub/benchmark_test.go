package hub_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/centrifugal/centrifuge"
	centrifugeClient "github.com/centrifugal/centrifuge-go"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gocasters/rankr/hub/emqx"
	"github.com/golang-jwt/jwt/v5"
)

const (
	emqxBroker   = "tcp://localhost:1883"
	emqxClientID = "benchmark-client"

	centrifugoURL    = "ws://localhost:8000/connection/websocket"
	centrifugoSecret = "secret"
)

type BenchmarkMessage struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   string    `json:"payload"`
}

func generatePayload(size int) string {
	payload := make([]byte, size)
	for i := range payload {
		payload[i] = 'a'
	}
	return string(payload)
}

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(centrifugoSecret))
}

func BenchmarkEMQX_ConnectionSetup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clientID := fmt.Sprintf("bench-conn-%d", i)
		client := emqx.NewClient(emqxBroker, clientID, nil)
		client.Disconnect()
	}
}

func BenchmarkEMQX_PublishSmallMessage(b *testing.B) {
	client := emqx.NewClient(emqxBroker, emqxClientID, nil)
	defer client.Disconnect()

	payload := generatePayload(100)
	topic := "benchmark/small"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_ = client.Publish(topic, 0, false, data)
	}
}

func BenchmarkEMQX_PublishMediumMessage(b *testing.B) {
	client := emqx.NewClient(emqxBroker, emqxClientID, nil)
	defer client.Disconnect()

	payload := generatePayload(1024)
	topic := "benchmark/medium"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_ = client.Publish(topic, 0, false, data)
	}
}

func BenchmarkEMQX_PublishLargeMessage(b *testing.B) {
	client := emqx.NewClient(emqxBroker, emqxClientID, nil)
	defer client.Disconnect()

	payload := generatePayload(10240)
	topic := "benchmark/large"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_ = client.Publish(topic, 0, false, data)
	}
}

func BenchmarkEMQX_PubSub_Latency(b *testing.B) {
	publisher := emqx.NewClient(emqxBroker, "bench-pub", nil)
	defer publisher.Disconnect()

	var receivedCount atomic.Int64
	doneChan := make(chan struct{}, b.N)

	handler := func(client mqtt.Client, msg mqtt.Message) {
		receivedCount.Add(1)
		doneChan <- struct{}{}
	}

	subscriber := emqx.NewClient(emqxBroker, "bench-sub", handler)
	defer subscriber.Disconnect()

	topic := "benchmark/latency"
	if err := subscriber.Subscribe(topic, 0, handler); err != nil {
		b.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "latency-test",
		}
		data, _ := json.Marshal(msg)
		_ = publisher.Publish(topic, 0, false, data)
		<-doneChan
	}

	b.StopTimer()
	b.ReportMetric(float64(receivedCount.Load()), "messages")
}

func BenchmarkEMQX_ConcurrentPublishers(b *testing.B) {
	workerCounts := []int{1, 10, 50, 100, 500}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			var wg sync.WaitGroup
			payload := generatePayload(100)

			b.ReportAllocs()
			b.ResetTimer()

			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					clientID := fmt.Sprintf("bench-worker-%d", workerID)
					client := emqx.NewClient(emqxBroker, clientID, nil)
					defer client.Disconnect()

					topic := fmt.Sprintf("benchmark/worker/%d", workerID)

					for i := 0; i < b.N/workers; i++ {
						msg := BenchmarkMessage{
							ID:        i,
							Timestamp: time.Now(),
							Payload:   payload,
						}
						data, _ := json.Marshal(msg)
						_ = client.Publish(topic, 0, false, data)
					}
				}(w)
			}

			wg.Wait()
		})
	}
}

func BenchmarkEMQX_Throughput(b *testing.B) {
	client := emqx.NewClient(emqxBroker, "bench-throughput", nil)
	defer client.Disconnect()

	payload := generatePayload(100)
	topic := "benchmark/throughput"

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_ = client.Publish(topic, 0, false, data)
	}

	elapsed := time.Since(start)
	throughput := float64(b.N) / elapsed.Seconds()

	b.ReportMetric(throughput, "msgs/sec")
}

func BenchmarkCentrifugo_ConnectionSetup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		token, err := generateToken(fmt.Sprintf("user-%d", i))
		if err != nil {
			b.Fatal(err)
		}

		client := centrifugeClient.NewJsonClient(
			centrifugoURL,
			centrifugeClient.Config{
				Token: token,
			},
		)

		if err := client.Connect(); err != nil {
			b.Fatal(err)
		}

		client.Disconnect()
	}
}

func BenchmarkCentrifugo_PublishSmallMessage(b *testing.B) {
	token, err := generateToken("bench-user")
	if err != nil {
		b.Fatal(err)
	}

	client := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: token,
		},
	)

	if err := client.Connect(); err != nil {
		b.Fatal(err)
	}
	defer client.Disconnect()

	channel := "benchmark:small"
	payload := generatePayload(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_, _ = client.Publish(context.Background(), channel, data)
	}
}

func BenchmarkCentrifugo_PublishMediumMessage(b *testing.B) {
	token, err := generateToken("bench-user")
	if err != nil {
		b.Fatal(err)
	}

	client := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: token,
		},
	)

	if err := client.Connect(); err != nil {
		b.Fatal(err)
	}
	defer client.Disconnect()

	channel := "benchmark:medium"
	payload := generatePayload(1024)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_, _ = client.Publish(context.Background(), channel, data)
	}
}

func BenchmarkCentrifugo_PublishLargeMessage(b *testing.B) {
	token, err := generateToken("bench-user")
	if err != nil {
		b.Fatal(err)
	}

	client := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: token,
		},
	)

	if err := client.Connect(); err != nil {
		b.Fatal(err)
	}
	defer client.Disconnect()

	channel := "benchmark:large"
	payload := generatePayload(10240)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_, _ = client.Publish(context.Background(), channel, data)
	}
}

func BenchmarkCentrifugo_PubSub_Latency(b *testing.B) {

	pubToken, _ := generateToken("bench-pub")
	publisher := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: pubToken,
		},
	)
	if err := publisher.Connect(); err != nil {
		b.Fatal(err)
	}
	defer publisher.Disconnect()

	subToken, _ := generateToken("bench-sub")
	subscriber := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: subToken,
		},
	)
	if err := subscriber.Connect(); err != nil {
		b.Fatal(err)
	}
	defer subscriber.Disconnect()

	channel := "benchmark:latency"
	var receivedCount atomic.Int64
	doneChan := make(chan struct{}, b.N)

	sub, err := subscriber.NewSubscription(channel)
	if err != nil {
		b.Fatal(err)
	}

	sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
		receivedCount.Add(1)
		doneChan <- struct{}{}
	})

	if err := sub.Subscribe(); err != nil {
		b.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "latency-test",
		}
		data, _ := json.Marshal(msg)
		_, _ = publisher.Publish(context.Background(), channel, data)
		<-doneChan
	}

	b.StopTimer()
	b.ReportMetric(float64(receivedCount.Load()), "messages")
}

func BenchmarkCentrifugo_ConcurrentPublishers(b *testing.B) {
	workerCounts := []int{1, 10, 50, 100, 500}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			var wg sync.WaitGroup
			payload := generatePayload(100)

			b.ReportAllocs()
			b.ResetTimer()

			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					token, _ := generateToken(fmt.Sprintf("worker-%d", workerID))
					client := centrifugeClient.NewJsonClient(
						centrifugoURL,
						centrifugeClient.Config{
							Token: token,
						},
					)

					if err := client.Connect(); err != nil {
						b.Error(err)
						return
					}
					defer client.Disconnect()

					channel := fmt.Sprintf("benchmark:worker:%d", workerID)

					for i := 0; i < b.N/workers; i++ {
						msg := BenchmarkMessage{
							ID:        i,
							Timestamp: time.Now(),
							Payload:   payload,
						}
						data, _ := json.Marshal(msg)
						_, _ = client.Publish(context.Background(), channel, data)
					}
				}(w)
			}

			wg.Wait()
		})
	}
}

func BenchmarkCentrifugo_Throughput(b *testing.B) {
	token, _ := generateToken("bench-throughput")
	client := centrifugeClient.NewJsonClient(
		centrifugoURL,
		centrifugeClient.Config{
			Token: token,
		},
	)

	if err := client.Connect(); err != nil {
		b.Fatal(err)
	}
	defer client.Disconnect()

	channel := "benchmark:throughput"
	payload := generatePayload(100)

	b.ReportAllocs()
	b.ResetTimer()

	start := time.Now()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_, _ = client.Publish(context.Background(), channel, data)
	}

	elapsed := time.Since(start)
	throughput := float64(b.N) / elapsed.Seconds()

	b.ReportMetric(throughput, "msgs/sec")
}

func BenchmarkComparison_MessageSerialization(b *testing.B) {
	payload := generatePayload(100)

	b.Run("JSON_Marshal", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			msg := BenchmarkMessage{
				ID:        i,
				Timestamp: time.Now(),
				Payload:   payload,
			}
			_, _ = json.Marshal(msg)
		}
	})

	b.Run("JSON_Unmarshal", func(b *testing.B) {
		msg := BenchmarkMessage{
			ID:        1,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var decoded BenchmarkMessage
			_ = json.Unmarshal(data, &decoded)
		}
	})
}

func BenchmarkComparison_MemoryAllocations(b *testing.B) {
	b.Run("EMQX_MessageCreation", func(b *testing.B) {
		b.ReportAllocs()
		payload := generatePayload(100)

		for i := 0; i < b.N; i++ {
			msg := BenchmarkMessage{
				ID:        i,
				Timestamp: time.Now(),
				Payload:   payload,
			}
			_, _ = json.Marshal(msg)
		}
	})

	b.Run("Centrifugo_MessageCreation", func(b *testing.B) {
		b.ReportAllocs()
		payload := generatePayload(100)

		for i := 0; i < b.N; i++ {
			msg := BenchmarkMessage{
				ID:        i,
				Timestamp: time.Now(),
				Payload:   payload,
			}
			_, _ = json.Marshal(msg)
		}
	})
}

func BenchmarkCentrifugo_ServerNode_Publish(b *testing.B) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		b.Fatal(err)
	}

	if err := node.Run(); err != nil {
		b.Fatal(err)
	}
	defer func() { _ = node.Shutdown(context.Background()) }()

	channel := "benchmark:server"
	payload := generatePayload(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   payload,
		}
		data, _ := json.Marshal(msg)
		_, _ = node.Publish(channel, data)
	}
}

func BenchmarkCentrifugo_ServerNode_ConcurrentPublish(b *testing.B) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		b.Fatal(err)
	}

	if err := node.Run(); err != nil {
		b.Fatal(err)
	}
	defer func() { _ = node.Shutdown(context.Background()) }()

	workerCounts := []int{1, 10, 50, 100}
	payload := generatePayload(100)

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("Workers_%d", workers), func(b *testing.B) {
			var wg sync.WaitGroup

			b.ReportAllocs()
			b.ResetTimer()

			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					channel := fmt.Sprintf("benchmark:server:%d", workerID)

					for i := 0; i < b.N/workers; i++ {
						msg := BenchmarkMessage{
							ID:        i,
							Timestamp: time.Now(),
							Payload:   payload,
						}
						data, _ := json.Marshal(msg)
						_, _ = node.Publish(channel, data)
					}
				}(w)
			}

			wg.Wait()
		})
	}
}
