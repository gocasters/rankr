package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	wnats "github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/labstack/echo/v4"
	nc "github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func (s *Server) TestSubscribe(c echo.Context) error {
	natsURL := "nats://supersecret@localhost:4222"
	ncConn, err := nc.Connect(natsURL, nc.Timeout(2*time.Second))
	if err != nil {
		fmt.Printf("Error connecting to NATS: %s\n", err.Error())
		return err
	}
	defer ncConn.Close()

	js, err := ncConn.JetStream()
	if err != nil {
		fmt.Printf("Error connecting to JetStream: %s\n", err.Error())
		return err
	}

	streamName := "EVENTS_STREAM"
	subject := "github.issues"
	durable := "durable_github_issues"

	// Pull subscription starting from startSeq
	sub, err := js.PullSubscribe(
		subject,
		durable,
		nc.BindStream(streamName),
	)
	if err != nil {
		fmt.Printf("Error pulling NATS subject: %s\n", err.Error())
		return err
	}

	// Fetch messages
	msgs, err := sub.Fetch(10, nc.MaxWait(5*time.Second))
	if err != nil {
		if errors.Is(err, nc.ErrTimeout) {
			return c.JSON(http.StatusOK, map[string]string{"message": "No messages returned"})
		}
		fmt.Printf("Error fetching messages from NATS: %s\n", err.Error())
		return err
	}

	marshaler := &wnats.GobMarshaler{}
	for _, msg := range msgs {
		wmMsg, err := marshaler.Unmarshal(msg)
		if err != nil {
			fmt.Printf("Error unmarshalling Watermill message: %v\n", err)
			// Decide whether to Ack/Nak based on your intent:
			if ackErr := msg.Ack(); ackErr != nil {
				fmt.Printf("Error acknowledging message: %v\n", ackErr)
			}

			// OR
			//if ackErr := msg.Nak(); ackErr != nil {
			//	fmt.Printf("Error nack message: %v\n", ackErr)
			//}
			continue
		}

		// If payload is protobuf, unmarshal it:
		var ev eventpb.Event
		if perr := proto.Unmarshal(wmMsg.Payload, &ev); perr != nil {
			fmt.Printf("Error unmarshalling payload proto: %v\n", perr)
		} else {
			fmt.Printf("Event: uuid: %s & payload: %+v\n", wmMsg.UUID, ev.Payload)
		}

		if ackErr := msg.Ack(); ackErr != nil {
			fmt.Printf("Error acknowledging message: %v\n", ackErr)
		}
	}

	return c.String(http.StatusOK, fmt.Sprintf("Printed %d messages", len(msgs)))
}
