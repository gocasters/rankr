package middleware

import "github.com/gocasters/rankr/contributorapp/client"

type Middleware struct {
	Config Config
	Client client.AuthClient
}

func New(cfg Config, client client.AuthClient) Middleware {
	return Middleware{Config: cfg, Client: client}
}
