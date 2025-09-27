package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gocasters/rankr/hub/emqx"
)

func main() {
	client := emqx.NewClient("tcp://localhost:1883", "go-client",
		func(c mqtt.Client, msg mqtt.Message) {
			fmt.Printf("📥 Received [%s]: %s\n", msg.Topic(), msg.Payload())
		})

	client.Subscribe("demo/leaderboard", 0, nil)

	select {} // keep running
}
