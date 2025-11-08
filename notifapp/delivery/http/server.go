package http

import (
	"context"
	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer *httpserver.Server
	handler    Handler
}

func NewServer(httpServer *httpserver.Server, handler Handler) Server {
	return Server{HTTPServer: httpServer, handler: handler}
}

func (s Server) Serve() error {
	s.registerRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) registerRoutes() {

	health := s.HTTPServer.GetRouter()

	health.GET("/v1/notifapp/health-check", s.healthCheck)

	v1 := s.HTTPServer.GetRouter().Group("/v1/notifapp/notifications", echomiddleware.ParseUserDataMiddleware)

	v1.GET("/:notification_id", s.handler.getNotification)
	v1.GET("", s.handler.listNotification)
	v1.GET("/unread/count", s.handler.getUnreadCount)

	v1.POST("", s.handler.createNotification)
	v1.PUT("/:notification_id", s.handler.markAsRead)
	v1.PUT("", s.handler.markAllAsRead)
	v1.DELETE("/:notification_id", s.handler.deleteNotification)
}
