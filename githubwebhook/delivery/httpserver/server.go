package httpserver

import (
	"context"

	"github.com/gocasters/rankr/githubwebhook/service"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer *httpserver.Server
	Handler    *Handler
	Service    *service.Service
}

func New(server *httpserver.Server, handler *Handler, svc *service.Service) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		Service:    svc,
	}
}

func (s *Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Server) stop(ctx context.Context) error {
	return s.HTTPServer.StopWithTimeout()
}

func (s *Server) RegisterRoutes() {
	webhookRouter := s.HTTPServer.GetRouter().Group("/github-webhook")

	webhookRouter.GET("/health-check", s.healthCheck)

	webhookRouter.POST("/", s.PublishGithubActivity)
}
