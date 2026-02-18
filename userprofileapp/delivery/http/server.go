package http

import (
	"context"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer *httpserver.Server
	Handler    *Handler
}

func New(httpServer *httpserver.Server, handler *Handler) Server {
	return Server{HTTPServer: httpServer, Handler: handler}
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
	router := s.HTTPServer.GetRouter()

	v1 := router.Group("/v1/userprofile")

	v1.GET("/health_check", s.healthCheck)

	v1.GET("/profile/:id", s.profile)
}
