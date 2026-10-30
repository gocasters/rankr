package userprofileapp

import (
	"fmt"

	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type Config struct {
	HTTPServer         httpserver.Config  `koanf:"http_server"`
	ContributorRPC     grpc.ClientConfig  `koanf:"contributor_grpc_client"`
	LeaderboardStatRPC grpc.ClientConfig  `koanf:"leaderboardstat_grpc_client"`
	Logger             logger.Config      `koanf:"logger"`
	Service            userprofile.Config `koanf:"service"`
}

func (c Config) Validate() error {
	if c.Service.UpstreamTimeout <= 0 {
		return fmt.Errorf("service.upstream_time_out must be greater than zero")
	}
	return nil
}
