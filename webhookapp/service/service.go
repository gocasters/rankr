package service

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type Service struct {
	Publisher message.Publisher
}

func New(publisher message.Publisher) *Service {
	return &Service{
		Publisher: publisher,
	}
}
