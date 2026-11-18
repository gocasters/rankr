package http

import (
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer *httpserver.Server
	Handler    Handler
}

func New(server *httpserver.Server, handler Handler) Server {
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

func (s Server) RegisterRoutes() {
	router := s.HTTPServer.GetRouter()

	v1 := router.Group("/v1")
	v1.GET("/health-check", s.healthCheck)
	v1.GET("/contributors/:user_id/activity", s.Handler.GetContributorActivity)
}
