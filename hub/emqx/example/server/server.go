package main

import (
	"fmt"
	"time"

	"github.com/gocasters/rankr/hub/emqx"
)

func main() {
	// Connect to EMQX broker
	server := emqx.NewClient("tcp://localhost:1883", "go-server", nil)
	defer server.Disconnect()

	fmt.Println("Server started: publishing data every 2 seconds...")

	counter := 0
	for {
		counter++
		msg := fmt.Sprintf("Real-time update #%d", counter)
		if err := server.Publish("demo/leaderboard", 0, false, msg); err != nil {
			fmt.Println("Publish error:", err)
		} else {
			fmt.Println("📤 Published:", msg)
		}
		time.Sleep(2 * time.Second)
	}
}
