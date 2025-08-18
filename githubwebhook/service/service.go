package service

import (
	"github.com/gocasters/rankr/event"
)

type Service struct {
	Publisher event.Publisher
}

func New(publisher event.Publisher) *Service {
	return &Service{
		Publisher: publisher,
	}
}
