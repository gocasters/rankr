package main

import (
	"encoding/json"
	"fmt"
	"log"

	centrifuge "github.com/centrifugal/centrifuge-go"
	"github.com/rivo/tview"
)

// Message structure
type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
}

func main() {
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() { app.Draw() })

	input := tview.NewInputField().SetLabel("Message: ")

	// Layout: text view on top, input at bottom
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(input, 1, 0, true)

	// Centrifugo connection
	client := centrifuge.NewJsonClient("ws://localhost:8000/connection/websocket", centrifuge.Config{})

	if err := client.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	fmt.Fprintln(textView, "[green]✅ Connected to Centrifugo")

	// Subscribe to channel
	channel := "chat"
	sub, err := client.NewSubscription(channel)
	if err != nil {
		log.Fatal("Subscription error:", err)
	}

	sub.OnPublication(func(pe centrifuge.PublicationEvent) {
		var msg Message
		if err := json.Unmarshal(pe.Data, &msg); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(textView, "[red]JSON parse error: %v\n", err)
			})
			return
		}
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(textView, "[yellow]%s: [white]%s\n", msg.User, msg.Text)
		})
	})

	if err := sub.Subscribe(); err != nil {
		log.Fatal("Subscribe failed:", err)
	}

	// Handle input sending
	// input.SetDoneFunc(func(key tview.Key) {
	// 	if key == tview.KeyEnter {
	// 		msg := Message{
	// 			User: "cli-user",
	// 			Text: input.GetText(),
	// 		}
	// 		data, _ := json.Marshal(msg)
	// 		if _, err := client.Publish(context.Background(), channel, data); err != nil {
	// 			fmt.Fprintf(textView, "[red]Publish error: %v\n", err)
	// 		}
	// 		input.SetText("")
	// 	}
	// })

	// Run TUI application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}
