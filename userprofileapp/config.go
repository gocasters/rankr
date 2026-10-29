package userprofileapp

import (
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
