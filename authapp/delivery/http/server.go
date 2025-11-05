package http

import (
	"context"

	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
}

func New(server httpserver.Server, handler Handler) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}
	return nil
}

func (s Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) RegisterRoutes() {
	v1 := s.HTTPServer.GetRouter().Group("v1")

	v1.GET("/health-check", s.Handler.healthCheck)
	v1.GET("/grant/:id", s.Handler.getGrant)
	v1.GET("/grants", s.Handler.listGrants)
	v1.POST("/grant", s.Handler.createGrant)
	v1.PUT("/grant", s.Handler.updateGrant)
	v1.DELETE("/grant/:id", s.Handler.deleteGrant)
}
