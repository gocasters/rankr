package http

import (
	"context"
	httpserver "github.com/gocasters/rankr/pkg/httpserver"
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
	// TODO- Add tracing middleware

	// create v1 group
	v1 := s.HTTPServer.GetRouter().Group("v1")
	v1.GET("/health-check", s.healthCheck)

	// contributor group
	contributorGroup := v1.Group("/contributors")
	contributorGroup.GET("/:id/stats", s.Handler.GetContributorStats)
	//contributorGroup.GET("/:id/rank", s.Handler.GetContributorRank)

}
