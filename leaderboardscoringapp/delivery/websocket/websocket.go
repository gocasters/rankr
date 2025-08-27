package websocket

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Host                string        `koanf:"host"`
	Port                int           `koanf:"port"`
	WebSocketPattern    string        `koanf:"websocket_pattern"`
	AllowedOrigins      []string      `koanf:"allowed_origins_websocket"`
	SendBufferSize      uint          `koanf:"send_buffer_size"`
	BroadcastBufferSize uint          `koanf:"broadcast_buffer_size"`
	ShutdownTimeout     time.Duration `koanf:"shutdown_context_timeout"`
}

type WebSocket struct {
	config     Config
	Hub        *Hub
	HTTPServer *httpserver.Server
	Logger     *slog.Logger
}

func New(cfg Config, logger *slog.Logger) (*WebSocket, error) {
	httpServer, err := httpserver.New(httpserver.Config{
		Port:            cfg.Port,
		CORS:            httpserver.CORS{AllowOrigins: cfg.AllowedOrigins},
		ShutdownTimeout: cfg.ShutdownTimeout,
		HideBanner:      true,
		HidePort:        true,
		OtelMiddleware:  nil,
	})
	if err != nil {
		return nil, err
	}

	return &WebSocket{
		config:     cfg,
		Hub:        NewHub(cfg.BroadcastBufferSize),
		HTTPServer: httpServer,
		Logger:     logger,
	}, nil
}

func (ws *WebSocket) Serve() error {
	v1 := ws.HTTPServer.GetRouter().Group("/v1")
	ldGroup := v1.Group("/leaderboard")
	ldGroup.GET(ws.config.WebSocketPattern, ws.socketHandler)

	go ws.Hub.Run()

	ws.Logger.Info(fmt.Sprintf("WebSocket started on %s:%d", ws.config.Host, ws.config.Port))
	if err := ws.HTTPServer.Start(); err != nil {
		return err
	}

	return nil
}

func (ws *WebSocket) Stop(ctx context.Context) error {
	return ws.HTTPServer.Stop(ctx)
}

func (ws *WebSocket) socketHandler(ctx echo.Context) error {

	r := ctx.Request()
	w := ctx.Response()

	upgrade := ws.NewUpgrader()
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		ws.Logger.Error("WebSocket upgrade failed", slog.String("error", err.Error()))
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "can't open websocket connection"})
	}

	client := &Client{
		Hub:  ws.Hub,
		Conn: conn,
		Send: make(chan []byte, ws.config.SendBufferSize),
		UUID: uuid.New(),
	}

	ws.Hub.GetRegisterChan() <- client

	go client.WritePump()
	go client.ReadPump()

	return nil
}

func (ws *WebSocket) NewUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
			origin := r.Header.Get("Origin")
			for _, allowed := range ws.config.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}
}
