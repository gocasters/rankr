package realtime

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan []byte
	Subscriptions map[string]bool
	SubsMu        sync.RWMutex
	ConnectedAt   time.Time
	LastActiveAt  time.Time
}

type Event struct {
	Type      string                 `json:"type"`
	Topic     string                 `json:"topic"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

type Message struct {
	Type    string                 `json:"type"`
	Topic   string                 `json:"topic,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type ErrorMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
