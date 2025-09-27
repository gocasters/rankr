package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centrifugal/centrifuge"
)

type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time string `json:"time"`
}

func auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cred := &centrifuge.Credentials{
			UserID: "anonymous", // Give a proper user ID
		}
		newCtx := centrifuge.SetCredentials(ctx, cred)
		r = r.WithContext(newCtx)
		h.ServeHTTP(w, r)
	})
}

func main() {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		log.Fatal(err)
	}

	node.OnConnect(func(client *centrifuge.Client) {
		transportName := client.Transport().Name()
		transportProto := client.Transport().Protocol()
		log.Printf("client connected via %s (%s)", transportName, transportProto)

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			log.Printf("client subscribes on channel %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, nil)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			log.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			log.Printf("client disconnected")
		})
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}

	// Start a goroutine to publish test messages every 10 seconds
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				msg := Message{
					User: "Server Bot",
					Text: "Hello from server! Time: " + time.Now().Format("15:04:05"),
					Time: time.Now().Format("2006-01-02 15:04:05"),
				}
				
				data, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error marshaling message: %v", err)
					continue
				}
				
				_, err = node.Publish( "chat", data)
				if err != nil {
					log.Printf("Error publishing message: %v", err)
				} else {
					log.Printf("Published message: %s", msg.Text)
				}
			}
		}
	}()

	// HTTP API endpoint to publish messages
	http.HandleFunc("/api/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		msg.Time = time.Now().Format("2006-01-02 15:04:05")
		data, _ := json.Marshal(msg)

		_, err := node.Publish( "chat", data)
		if err != nil {
			http.Error(w, "Failed to publish", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{})
	http.Handle("/connection/websocket", auth(wsHandler))
	http.Handle("/", http.FileServer(http.Dir("./")))

	log.Printf("Starting server on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}