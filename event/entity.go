package event

import "github.com/gocasters/rankr/protobuf/golang/eventpb"

type Topic string

type Event struct {
	Topic   Topic
	Payload []byte
}

type Handler func(event *eventpb.Event) error