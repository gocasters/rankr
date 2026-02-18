package http

import (
	"context"
	authmiddleware "github.com/gocasters/rankr/authapp/delivery/http/middleware"
	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	Middleware authmiddleware.Middleware
}

func New(server httpserver.Server, handler Handler) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		Middleware: authmiddleware.New(handler.tokenService),
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
	router := s.HTTPServer.GetRouter()
	router.Use(
		s.Middleware.RequireBearerToken(
			authmiddleware.RequireBearerTokenOptions{
				Skipper: echomiddleware.SkipExactPaths(
					"/v1/health-check",
					"/v1/login",
					"/v1/refresh-token",
				),
			},
		),
	)

	v1 := router.Group("v1")

	v1.GET("/health-check", s.Handler.healthCheck)
	v1.POST("/login", s.Handler.login)
	v1.POST("/refresh-token", s.Handler.refreshToken)
	v1.GET("/me", s.Handler.me)
}
