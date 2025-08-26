package eventpublisher

import (
	"context"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
)

type Publisher interface {
	Publish(ctx context.Context, evnet *eventpb.Event) error
}
