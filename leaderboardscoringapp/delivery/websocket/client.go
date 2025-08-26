package websocket

import (
	"github.com/google/uuid"
)

type WebSocketConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type Client struct {
	Hub  HubInterface
	Conn WebSocketConn
	Send chan []byte
	UUID uuid.UUID
}

// ReadPump TODO -
func (c *Client) ReadPump() {
	c.Conn.ReadMessage()
}

// WritePump TODO - Write data to websocket
func (c *Client) WritePump() {}

// Close TODO - Safely closes connection and Unregisters client
func (c *Client) Close() {}
