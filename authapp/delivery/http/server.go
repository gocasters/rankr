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
	v1.POST("/login", s.Handler.login)
	v1.POST("/token/verify", s.Handler.verifyToken)
	v1.GET("/role/:id", s.Handler.getRole)
	v1.GET("/roles", s.Handler.listRoles)
	v1.POST("/role", s.Handler.createRole)
	v1.PUT("/role", s.Handler.updateRole)
	v1.DELETE("/role/:id", s.Handler.deleteRole)
	v1.POST("/role/permission", s.Handler.addPermissionToRole)
	v1.DELETE("/role/permission", s.Handler.removePermissionFromRole)
}
