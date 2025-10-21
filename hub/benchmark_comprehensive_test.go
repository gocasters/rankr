package hub_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	centrifugeClient "github.com/centrifugal/centrifuge-go"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gocasters/rankr/hub/emqx"
)

type LeaderboardUpdate struct {
	UserID    string    `json:"user_id"`
	Score     int       `json:"score"`
	Rank      int       `json:"rank"`
	Timestamp time.Time `json:"timestamp"`
}

type TaskNotification struct {
	TaskID      string    `json:"task_id"`
	Action      string    `json:"action"`
	AssignedTo  string    `json:"assigned_to"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

type ContributorActivity struct {
	ContributorID string    `json:"contributor_id"`
	ActivityType  string    `json:"activity_type"`
	ProjectID     string    `json:"project_id"`
	Timestamp     time.Time `json:"timestamp"`
}

func BenchmarkEMQX_Scenario_Leaderboard(b *testing.B) {
	subscriberCounts := []int{10, 50, 100}

	for _, subCount := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers_%d", subCount), func(b *testing.B) {
			topic := fmt.Sprintf("leaderboard/updates/%d", subCount)
			var receivedCount atomic.Int64
			var wg sync.WaitGroup

			subscribers := make([]*emqx.MQTTClient, subCount)
			for i := 0; i < subCount; i++ {
				clientID := fmt.Sprintf("sub-%d-%d", subCount, i)
				handler := func(client mqtt.Client, msg mqtt.Message) {
					receivedCount.Add(1)
				}
				subscribers[i] = emqx.NewClient(emqxBroker, clientID, handler)
				_ = subscribers[i].Subscribe(topic, 0, handler)
			}
			defer func() {
				for _, sub := range subscribers {
					sub.Disconnect()
				}
			}()

			time.Sleep(100 * time.Millisecond)

			publisher := emqx.NewClient(emqxBroker, fmt.Sprintf("pub-%d", subCount), nil)
			defer publisher.Disconnect()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				update := LeaderboardUpdate{
					UserID:    fmt.Sprintf("user-%d", i),
					Score:     rand.Intn(10000),
					Rank:      i + 1,
					Timestamp: time.Now(),
				}
				data, _ := json.Marshal(update)
				_ = publisher.Publish(topic, 0, false, data)
			}

			wg.Wait()
			b.StopTimer()

			expectedMessages := int64(b.N * subCount)
			time.Sleep(500 * time.Millisecond)
			received := receivedCount.Load()

			b.ReportMetric(float64(received), "msgs_delivered")
			b.ReportMetric(float64(received)/float64(expectedMessages)*100, "delivery_%")
		})
	}
}

func BenchmarkCentrifugo_Scenario_Leaderboard(b *testing.B) {
	subscriberCounts := []int{10, 50, 100}

	for _, subCount := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers_%d", subCount), func(b *testing.B) {
			channel := fmt.Sprintf("leaderboard:updates:%d", subCount)
			var receivedCount atomic.Int64

			subscribers := make([]*centrifugeClient.Client, subCount)
			subscriptions := make([]*centrifugeClient.Subscription, subCount)

			for i := 0; i < subCount; i++ {
				token, _ := generateToken(fmt.Sprintf("sub-%d-%d", subCount, i))
				client := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
				_ = client.Connect()

				sub, _ := client.NewSubscription(channel)
				sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
					receivedCount.Add(1)
				})
				_ = sub.Subscribe()

				subscribers[i] = client
				subscriptions[i] = sub
			}
			defer func() {
				for _, client := range subscribers {
					client.Disconnect()
				}
			}()

			time.Sleep(100 * time.Millisecond)

			pubToken, _ := generateToken(fmt.Sprintf("pub-%d", subCount))
			publisher := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: pubToken})
			_ = publisher.Connect()
			defer publisher.Disconnect()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				update := LeaderboardUpdate{
					UserID:    fmt.Sprintf("user-%d", i),
					Score:     rand.Intn(10000),
					Rank:      i + 1,
					Timestamp: time.Now(),
				}
				data, _ := json.Marshal(update)
				_, _ = publisher.Publish(context.Background(), channel, data)
			}

			b.StopTimer()

			expectedMessages := int64(b.N * subCount)
			time.Sleep(500 * time.Millisecond)
			received := receivedCount.Load()

			b.ReportMetric(float64(received), "msgs_delivered")
			b.ReportMetric(float64(received)/float64(expectedMessages)*100, "delivery_%")
		})
	}
}

func BenchmarkEMQX_Scenario_MultiChannel(b *testing.B) {
	channels := []string{
		"leaderboard/updates",
		"tasks/notifications",
		"contributor/activity",
		"project/events",
	}

	var receivedCount atomic.Int64
	handler := func(client mqtt.Client, msg mqtt.Message) {
		receivedCount.Add(1)
	}

	client := emqx.NewClient(emqxBroker, "multi-channel-client", handler)
	defer client.Disconnect()

	for _, channel := range channels {
		_ = client.Subscribe(channel, 0, handler)
	}

	time.Sleep(100 * time.Millisecond)

	publisher := emqx.NewClient(emqxBroker, "multi-channel-pub", nil)
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		channel := channels[i%len(channels)]
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "test",
		}
		data, _ := json.Marshal(msg)
		_ = publisher.Publish(channel, 0, false, data)
	}

	b.StopTimer()
	time.Sleep(200 * time.Millisecond)
	b.ReportMetric(float64(receivedCount.Load()), "messages")
}

func BenchmarkCentrifugo_Scenario_MultiChannel(b *testing.B) {
	channels := []string{
		"leaderboard:updates",
		"tasks:notifications",
		"contributor:activity",
		"project:events",
	}

	var receivedCount atomic.Int64

	token, _ := generateToken("multi-channel-client")
	client := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
	_ = client.Connect()
	defer client.Disconnect()

	for _, channel := range channels {
		sub, _ := client.NewSubscription(channel)
		sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
			receivedCount.Add(1)
		})
		_ = sub.Subscribe()
	}

	time.Sleep(100 * time.Millisecond)

	pubToken, _ := generateToken("multi-channel-pub")
	publisher := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: pubToken})
	_ = publisher.Connect()
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		channel := channels[i%len(channels)]
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "test",
		}
		data, _ := json.Marshal(msg)
		_, _ = publisher.Publish(context.Background(), channel, data)
	}

	b.StopTimer()
	time.Sleep(200 * time.Millisecond)
	b.ReportMetric(float64(receivedCount.Load()), "messages")
}

func BenchmarkEMQX_Scenario_BurstTraffic(b *testing.B) {
	burstSizes := []int{10, 50, 100}

	for _, burstSize := range burstSizes {
		b.Run(fmt.Sprintf("Burst_%d", burstSize), func(b *testing.B) {
			client := emqx.NewClient(emqxBroker, "burst-client", nil)
			defer client.Disconnect()

			topic := fmt.Sprintf("burst/test/%d", burstSize)
			payload := generatePayload(100)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				for j := 0; j < burstSize; j++ {
					msg := BenchmarkMessage{
						ID:        i*burstSize + j,
						Timestamp: time.Now(),
						Payload:   payload,
					}
					data, _ := json.Marshal(msg)
					_ = client.Publish(topic, 0, false, data)
				}

				if i%10 == 0 {
					time.Sleep(10 * time.Millisecond)
				}
			}
		})
	}
}

func BenchmarkCentrifugo_Scenario_BurstTraffic(b *testing.B) {
	burstSizes := []int{10, 50, 100}

	for _, burstSize := range burstSizes {
		b.Run(fmt.Sprintf("Burst_%d", burstSize), func(b *testing.B) {
			token, _ := generateToken("burst-client")
			client := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
			_ = client.Connect()
			defer client.Disconnect()

			channel := fmt.Sprintf("burst:test:%d", burstSize)
			payload := generatePayload(100)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {

				for j := 0; j < burstSize; j++ {
					msg := BenchmarkMessage{
						ID:        i*burstSize + j,
						Timestamp: time.Now(),
						Payload:   payload,
					}
					data, _ := json.Marshal(msg)
					_, _ = client.Publish(context.Background(), channel, data)
				}

				if i%10 == 0 {
					time.Sleep(10 * time.Millisecond)
				}
			}
		})
	}
}

func BenchmarkEMQX_Scenario_ConnectionChurn(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		clientID := fmt.Sprintf("churn-%d", i)
		client := emqx.NewClient(emqxBroker, clientID, nil)

		_ = client.Subscribe("churn/test", 0, nil)

		_ = client.Publish("churn/test", 0, false, []byte("test"))

		client.Disconnect()

		if i%100 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func BenchmarkCentrifugo_Scenario_ConnectionChurn(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		token, _ := generateToken(fmt.Sprintf("churn-%d", i))
		client := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
		_ = client.Connect()

		sub, _ := client.NewSubscription("churn:test")
		_ = sub.Subscribe()

		_, _ = client.Publish(context.Background(), "churn:test", []byte("test"))

		client.Disconnect()

		if i%100 == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func BenchmarkEMQX_Scenario_MixedWorkload(b *testing.B) {
	ratios := []struct {
		name     string
		readPct  int
		writePct int
	}{
		{"Read_Heavy_80_20", 80, 20},
		{"Balanced_50_50", 50, 50},
		{"Write_Heavy_20_80", 20, 80},
	}

	for _, ratio := range ratios {
		b.Run(ratio.name, func(b *testing.B) {
			var receivedCount atomic.Int64

			subscriber := emqx.NewClient(emqxBroker, "mixed-sub", func(client mqtt.Client, msg mqtt.Message) {
				receivedCount.Add(1)
			})
			defer subscriber.Disconnect()
			_ = subscriber.Subscribe("mixed/workload", 0, func(client mqtt.Client, msg mqtt.Message) {
				receivedCount.Add(1)
			})

			publisher := emqx.NewClient(emqxBroker, "mixed-pub", nil)
			defer publisher.Disconnect()

			time.Sleep(100 * time.Millisecond)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				action := rand.Intn(100)

				if action < ratio.writePct {

					msg := BenchmarkMessage{
						ID:        i,
						Timestamp: time.Now(),
						Payload:   "test",
					}
					data, _ := json.Marshal(msg)
					_ = publisher.Publish("mixed/workload", 0, false, data)
				} else {

					time.Sleep(1 * time.Microsecond)
				}
			}

			b.StopTimer()
			time.Sleep(100 * time.Millisecond)
			b.ReportMetric(float64(receivedCount.Load()), "messages_received")
		})
	}
}

func BenchmarkCentrifugo_Scenario_MixedWorkload(b *testing.B) {
	ratios := []struct {
		name     string
		readPct  int
		writePct int
	}{
		{"Read_Heavy_80_20", 80, 20},
		{"Balanced_50_50", 50, 50},
		{"Write_Heavy_20_80", 20, 80},
	}

	for _, ratio := range ratios {
		b.Run(ratio.name, func(b *testing.B) {
			var receivedCount atomic.Int64

			subToken, _ := generateToken("mixed-sub")
			subscriber := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: subToken})
			_ = subscriber.Connect()
			defer subscriber.Disconnect()

			sub, _ := subscriber.NewSubscription("mixed:workload")
			sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
				receivedCount.Add(1)
			})
			_ = sub.Subscribe()

			pubToken, _ := generateToken("mixed-pub")
			publisher := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: pubToken})
			_ = publisher.Connect()
			defer publisher.Disconnect()

			time.Sleep(100 * time.Millisecond)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				action := rand.Intn(100)

				if action < ratio.writePct {

					msg := BenchmarkMessage{
						ID:        i,
						Timestamp: time.Now(),
						Payload:   "test",
					}
					data, _ := json.Marshal(msg)
					_, _ = publisher.Publish(context.Background(), "mixed:workload", data)
				} else {

					time.Sleep(1 * time.Microsecond)
				}
			}

			b.StopTimer()
			time.Sleep(100 * time.Millisecond)
			b.ReportMetric(float64(receivedCount.Load()), "messages_received")
		})
	}
}

func BenchmarkEMQX_Scenario_PeakLoad_Rankr(b *testing.B) {

	users := 100
	channels := []string{"leaderboard/updates", "tasks/notifications", "contributor/activity"}

	var receivedCount atomic.Int64
	var clients []*emqx.MQTTClient

	for i := 0; i < users; i++ {
		clientID := fmt.Sprintf("user-%d", i)
		handler := func(client mqtt.Client, msg mqtt.Message) {
			receivedCount.Add(1)
		}
		client := emqx.NewClient(emqxBroker, clientID, handler)

		for _, channel := range channels {
			_ = client.Subscribe(channel, 0, handler)
		}

		clients = append(clients, client)
	}
	defer func() {
		for _, client := range clients {
			client.Disconnect()
		}
	}()

	time.Sleep(200 * time.Millisecond)

	publishers := make([]*emqx.MQTTClient, len(channels))
	for i, channel := range channels {
		pubID := fmt.Sprintf("pub-%s", channel)
		publishers[i] = emqx.NewClient(emqxBroker, pubID, nil)
	}
	defer func() {
		for _, pub := range publishers {
			pub.Disconnect()
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		for idx, channel := range channels {
			var data []byte
			switch idx {
			case 0:
				update := LeaderboardUpdate{
					UserID:    fmt.Sprintf("user-%d", i%users),
					Score:     rand.Intn(10000),
					Rank:      i%users + 1,
					Timestamp: time.Now(),
				}
				data, _ = json.Marshal(update)
			case 1:
				notif := TaskNotification{
					TaskID:      fmt.Sprintf("task-%d", i),
					Action:      "updated",
					AssignedTo:  fmt.Sprintf("user-%d", i%users),
					Description: "Test task",
					Timestamp:   time.Now(),
				}
				data, _ = json.Marshal(notif)
			case 2:
				activity := ContributorActivity{
					ContributorID: fmt.Sprintf("contributor-%d", i%users),
					ActivityType:  "commit",
					ProjectID:     "rankr",
					Timestamp:     time.Now(),
				}
				data, _ = json.Marshal(activity)
			}
			_ = publishers[idx].Publish(channel, 0, false, data)
		}

		if i%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	b.StopTimer()
	time.Sleep(500 * time.Millisecond)

	expectedMessages := int64(b.N * len(channels) * users)
	received := receivedCount.Load()

	b.ReportMetric(float64(received), "total_messages")
	b.ReportMetric(float64(received)/float64(expectedMessages)*100, "delivery_%")
}

func BenchmarkCentrifugo_Scenario_PeakLoad_Rankr(b *testing.B) {
	users := 100
	channels := []string{"leaderboard:updates", "tasks:notifications", "contributor:activity"}

	var receivedCount atomic.Int64
	var clients []*centrifugeClient.Client

	for i := 0; i < users; i++ {
		token, _ := generateToken(fmt.Sprintf("user-%d", i))
		client := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
		_ = client.Connect()

		for _, channel := range channels {
			sub, _ := client.NewSubscription(channel)
			sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
				receivedCount.Add(1)
			})
			_ = sub.Subscribe()
		}

		clients = append(clients, client)
	}
	defer func() {
		for _, client := range clients {
			client.Disconnect()
		}
	}()

	time.Sleep(200 * time.Millisecond)

	publishers := make([]*centrifugeClient.Client, len(channels))
	for i := range channels {
		token, _ := generateToken(fmt.Sprintf("pub-%d", i))
		pub := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
		_ = pub.Connect()
		publishers[i] = pub
	}
	defer func() {
		for _, pub := range publishers {
			pub.Disconnect()
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		for idx, channel := range channels {
			var data []byte
			switch idx {
			case 0:
				update := LeaderboardUpdate{
					UserID:    fmt.Sprintf("user-%d", i%users),
					Score:     rand.Intn(10000),
					Rank:      i%users + 1,
					Timestamp: time.Now(),
				}
				data, _ = json.Marshal(update)
			case 1:
				notif := TaskNotification{
					TaskID:      fmt.Sprintf("task-%d", i),
					Action:      "updated",
					AssignedTo:  fmt.Sprintf("user-%d", i%users),
					Description: "Test task",
					Timestamp:   time.Now(),
				}
				data, _ = json.Marshal(notif)
			case 2:
				activity := ContributorActivity{
					ContributorID: fmt.Sprintf("contributor-%d", i%users),
					ActivityType:  "commit",
					ProjectID:     "rankr",
					Timestamp:     time.Now(),
				}
				data, _ = json.Marshal(activity)
			}
			_, _ = publishers[idx].Publish(context.Background(), channel, data)
		}

		if i%10 == 0 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	b.StopTimer()
	time.Sleep(500 * time.Millisecond)

	expectedMessages := int64(b.N * len(channels) * users)
	received := receivedCount.Load()

	b.ReportMetric(float64(received), "total_messages")
	b.ReportMetric(float64(received)/float64(expectedMessages)*100, "delivery_%")
}

func BenchmarkEMQX_Scenario_MessageOrdering(b *testing.B) {
	var receivedMessages []int
	var mu sync.Mutex

	handler := func(client mqtt.Client, msg mqtt.Message) {
		var bmsg BenchmarkMessage
		_ = json.Unmarshal(msg.Payload(), &bmsg)
		mu.Lock()
		receivedMessages = append(receivedMessages, bmsg.ID)
		mu.Unlock()
	}

	subscriber := emqx.NewClient(emqxBroker, "ordering-sub", handler)
	defer subscriber.Disconnect()
	_ = subscriber.Subscribe("ordering/test", 1, handler)

	time.Sleep(100 * time.Millisecond)

	publisher := emqx.NewClient(emqxBroker, "ordering-pub", nil)
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "test",
		}
		data, _ := json.Marshal(msg)
		_ = publisher.Publish("ordering/test", 1, false, data)
	}

	b.StopTimer()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	outOfOrder := 0
	for i := 1; i < len(receivedMessages); i++ {
		if receivedMessages[i] < receivedMessages[i-1] {
			outOfOrder++
		}
	}

	b.ReportMetric(float64(len(receivedMessages)), "received")
	b.ReportMetric(float64(outOfOrder), "out_of_order")
	b.ReportMetric(float64(len(receivedMessages))/float64(b.N)*100, "delivery_%")
}

func BenchmarkCentrifugo_Scenario_MessageOrdering(b *testing.B) {
	var receivedMessages []int
	var mu sync.Mutex

	token, _ := generateToken("ordering-sub")
	subscriber := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
	_ = subscriber.Connect()
	defer subscriber.Disconnect()

	sub, _ := subscriber.NewSubscription("ordering:test")
	sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
		var bmsg BenchmarkMessage
		_ = json.Unmarshal(e.Data, &bmsg)
		mu.Lock()
		receivedMessages = append(receivedMessages, bmsg.ID)
		mu.Unlock()
	})
	_ = sub.Subscribe()

	time.Sleep(100 * time.Millisecond)

	pubToken, _ := generateToken("ordering-pub")
	publisher := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: pubToken})
	_ = publisher.Connect()
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := BenchmarkMessage{
			ID:        i,
			Timestamp: time.Now(),
			Payload:   "test",
		}
		data, _ := json.Marshal(msg)
		_, _ = publisher.Publish(context.Background(), "ordering:test", data)
	}

	b.StopTimer()
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	outOfOrder := 0
	for i := 1; i < len(receivedMessages); i++ {
		if receivedMessages[i] < receivedMessages[i-1] {
			outOfOrder++
		}
	}

	b.ReportMetric(float64(len(receivedMessages)), "received")
	b.ReportMetric(float64(outOfOrder), "out_of_order")
	b.ReportMetric(float64(len(receivedMessages))/float64(b.N)*100, "delivery_%")
}

func BenchmarkEMQX_Scenario_SustainedLoad(b *testing.B) {
	duration := 10 * time.Second
	targetRate := 1000

	var sentCount atomic.Int64
	var receivedCount atomic.Int64

	handler := func(client mqtt.Client, msg mqtt.Message) {
		receivedCount.Add(1)
	}

	subscriber := emqx.NewClient(emqxBroker, "sustained-sub", handler)
	defer subscriber.Disconnect()
	_ = subscriber.Subscribe("sustained/test", 0, handler)

	time.Sleep(100 * time.Millisecond)

	publisher := emqx.NewClient(emqxBroker, "sustained-pub", nil)
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	ticker := time.NewTicker(time.Second / time.Duration(targetRate))
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			b.StopTimer()
			time.Sleep(500 * time.Millisecond)

			sent := sentCount.Load()
			received := receivedCount.Load()

			b.ReportMetric(float64(sent), "sent")
			b.ReportMetric(float64(received), "received")
			b.ReportMetric(float64(sent)/duration.Seconds(), "rate_msg/s")
			b.ReportMetric(float64(received)/float64(sent)*100, "delivery_%")
			return

		case <-ticker.C:
			msg := BenchmarkMessage{
				ID:        int(sentCount.Load()),
				Timestamp: time.Now(),
				Payload:   "sustained",
			}
			data, _ := json.Marshal(msg)
			_ = publisher.Publish("sustained/test", 0, false, data)
			sentCount.Add(1)
		}
	}
}

func BenchmarkCentrifugo_Scenario_SustainedLoad(b *testing.B) {
	duration := 10 * time.Second
	targetRate := 1000

	var sentCount atomic.Int64
	var receivedCount atomic.Int64

	token, _ := generateToken("sustained-sub")
	subscriber := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: token})
	_ = subscriber.Connect()
	defer subscriber.Disconnect()

	sub, _ := subscriber.NewSubscription("sustained:test")
	sub.OnPublication(func(e centrifugeClient.PublicationEvent) {
		receivedCount.Add(1)
	})
	_ = sub.Subscribe()

	time.Sleep(100 * time.Millisecond)

	pubToken, _ := generateToken("sustained-pub")
	publisher := centrifugeClient.NewJsonClient(centrifugoURL, centrifugeClient.Config{Token: pubToken})
	_ = publisher.Connect()
	defer publisher.Disconnect()

	b.ReportAllocs()
	b.ResetTimer()

	ticker := time.NewTicker(time.Second / time.Duration(targetRate))
	defer ticker.Stop()

	timeout := time.After(duration)

	for {
		select {
		case <-timeout:
			b.StopTimer()
			time.Sleep(500 * time.Millisecond)

			sent := sentCount.Load()
			received := receivedCount.Load()

			b.ReportMetric(float64(sent), "sent")
			b.ReportMetric(float64(received), "received")
			b.ReportMetric(float64(sent)/duration.Seconds(), "rate_msg/s")
			b.ReportMetric(float64(received)/float64(sent)*100, "delivery_%")
			return

		case <-ticker.C:
			msg := BenchmarkMessage{
				ID:        int(sentCount.Load()),
				Timestamp: time.Now(),
				Payload:   "sustained",
			}
			data, _ := json.Marshal(msg)
			_, _ = publisher.Publish(context.Background(), "sustained:test", data)
			sentCount.Add(1)
		}
	}
}
