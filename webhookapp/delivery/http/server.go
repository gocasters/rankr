package http

import (
	"context"
	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/webhookapp/service/delivery"
)

type Server struct {
	HTTPServer *httpserver.Server
	Handler    *Handler
	Service    *delivery.Service
}

func New(server *httpserver.Server, handler *Handler, svc *delivery.Service) Server {
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

func (s *Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s *Server) RegisterRoutes() {
	router := s.HTTPServer.GetRouter()
	router.Use(
		echomiddleware.RequireClaimsWithConfig(
			echomiddleware.RequireClaimsConfig{
				Skipper: echomiddleware.SkipExactPaths("/github-webhook/health-check"),
			},
		),
	)

	webhookRouter := router.Group("/github-webhook")

	webhookRouter.GET("/health-check", s.healthCheck)

	webhookRouter.POST("/github/process", s.PublishGithubActivity)
}
