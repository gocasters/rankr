package main

import (
	"time"

	"github.com/gocasters/rankr/hub/emqx"
)

func main() {
	server := emqx.NewClient("tcp://localhost:1883", "go-server", nil)
	server.PublishLoop("demo/leaderboard", 5, 2*time.Second)
	server.Disconnect()
}
