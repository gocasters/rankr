package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	centrifuge "github.com/centrifugal/centrifuge-go"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Message struct {
	User string `json:"user"`
	Text string `json:"text"`
	Time string `json:"time"`
}

func generateToken(userID, secret string) string {
	// Create JWT header
	header := map[string]interface{}{
		"typ": "JWT",
		"alg": "HS256",
	}

	// Create JWT payload
	payload := map[string]interface{}{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	// Encode header and payload
	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

func main() {
	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetChangedFunc(func() {
			app.QueueUpdateDraw(func() {})
		})

	input := tview.NewInputField().
		SetLabel("Message: ").
		SetFieldWidth(0)

	// Create layout
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(input, 1, 0, true)

	// Add initial message
	fmt.Fprintln(textView, "[blue]📱 Chat Client Starting...")

	// Generate auth token (use the same secret as in config.json)
	userID := "client-user-123"
	secret := "bbe7d157-a253-4094-9759-06a8236543f9"
	token := generateToken(userID, secret)

	// Centrifugo connection with authentication
	config := centrifuge.Config{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Name:         "go-client",
		Version:      "1.0.0",
		Token:        token,
	}

	client := centrifuge.NewJsonClient("ws://localhost:8000/connection/websocket", config)

	// Set connection handlers
	client.OnConnected(func(e centrifuge.ConnectedEvent) {
		app.QueueUpdateDraw(func() {
			fmt.Fprintln(textView, "[green]✅ Connected to Centrifugo")
		})
	})

	client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(textView, "[red]❌ Disconnected: %s\n", e.Reason)
		})
	})

	client.OnError(func(e centrifuge.ErrorEvent) {
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(textView, "[red]Error: %v\n", e.Error)
		})
	})

	// Connect to server
	if err := client.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Subscribe to channel
	channel := "chat"
	sub, err := client.NewSubscription(channel)
	if err != nil {
		log.Fatal("Subscription error:", err)
	}

	// Handle incoming messages
	sub.OnPublication(func(pe centrifuge.PublicationEvent) {
		var msg Message
		if err := json.Unmarshal(pe.Data, &msg); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(textView, "[red]JSON parse error: %v\n", err)
			})
			return
		}

		app.QueueUpdateDraw(func() {
			timeStr := ""
			if msg.Time != "" {
				timeStr = fmt.Sprintf(" [gray](%s)[white]", msg.Time)
			}
			fmt.Fprintf(textView, "[yellow]%s:[white] %s%s\n", msg.User, msg.Text, timeStr)
		})
	})

	sub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
		app.QueueUpdateDraw(func() {
			fmt.Fprintln(textView, "[green]📺 Subscribed to chat channel")
		})
	})

	sub.OnError(func(e centrifuge.SubscriptionErrorEvent) {
		app.QueueUpdateDraw(func() {
			fmt.Fprintf(textView, "[red]Subscription error: %v\n", e.Error)
		})
	})

	if err := sub.Subscribe(); err != nil {
		log.Fatal("Subscribe failed:", err)
	}

	// Handle input for sending messages
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			text := input.GetText()
			if text != "" {
				msg := Message{
					User: "TUI Client",
					Text: text,
					Time: time.Now().Format("2006-01-02 15:04:05"),
				}

				data, err := json.Marshal(msg)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(textView, "[red]Error marshaling message: %v\n", err)
					})
					return nil
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				_, err = client.Publish(ctx, channel, data)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(textView, "[red]Publish error: %v\n", err)
					})
				} else {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(textView, "[cyan]You: [white]%s\n", text)
					})
				}
				input.SetText("")
			}
			return nil
		}
		return event
	})

	// Set focus and run application
	app.SetFocus(input)

	// Handle graceful shutdown
	defer func() {
		if client != nil {
			client.Close()
		}
	}()

	if err := app.SetRoot(flex, true).Run(); err != nil {
		log.Fatal(err)
	}
}
