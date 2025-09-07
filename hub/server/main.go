package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gocasters/rankr/hub"
)

func main() {
	channel := "leaderboard"

	centrifugoURL := os.Getenv("CENTRIFUGO_URL")
	centrifugoAPIKey := os.Getenv("CENTRIFUGO_API_KEY")
	if centrifugoURL == "" || centrifugoAPIKey == "" {
		log.Fatal("CENTRIFUGO_URL and CENTRIFUGO_API_KEY must be set")
	}

	h := hub.NewCentrifugoHub(centrifugoURL, centrifugoAPIKey)

	http.HandleFunc("/score", func(w http.ResponseWriter, r *http.Request) {
		player := r.URL.Query().Get("player")
		score := r.URL.Query().Get("score")
		if player == "" || score == "" {
			http.Error(w, "player and score required", http.StatusBadRequest)
			return
		}

		msg := fmt.Sprintf("%s scored %s points", player, score)
		if err := h.Publish(context.Background(), channel, msg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "ok")
	})

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
