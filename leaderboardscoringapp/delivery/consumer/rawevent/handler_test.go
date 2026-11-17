package rawevent

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
)

// A payload that cannot be decoded or mapped fails the same way on every
// redelivery, so it is ACKed instead of looping. The raw event stays in
// webhook_db and can be replayed once the bug is fixed.
func TestHandleEventAcksUndecodablePayload(t *testing.T) {
	if err := logger.Init(logger.Config{}); err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name    string
		payload []byte
	}{
		{name: "invalid protobuf", payload: []byte{0xff}},
		{name: "unmappable event", payload: nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := message.NewMessage("test", tc.payload)
			if err := (Handler{}).HandleEvent(msg); err != nil {
				t.Fatalf("HandleEvent() = %v, want nil (ACK, no scoring write)", err)
			}
		})
	}
}

func TestHandleProcessingResult(t *testing.T) {
	if err := logger.Init(logger.Config{}); err != nil {
		t.Fatal(err)
	}
	transientErr := errors.New("redis unavailable")

	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{name: "successful write"},
		{name: "already processed", err: ErrEventAlreadyProcessed},
		{name: "locked", err: ErrEventLocked, want: ErrEventLocked},
		{name: "invalid event is ACKed", err: errors.Join(leaderboardscoring.ErrInvalidEventRequest, errors.New("timestamp: cannot be blank"))},
		{name: "transient failure", err: transientErr, want: transientErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := handleProcessingResult("event-1", tc.err)
			if got != tc.want {
				t.Errorf("handleProcessingResult() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestIsPermanent(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "joined invalid event error",
			err:  errors.Join(leaderboardscoring.ErrInvalidEventRequest, errors.New("timestamp: cannot be blank")),
			want: true,
		},
		{
			name: "wrapped invalid event error",
			err:  fmt.Errorf("process event: %w", leaderboardscoring.ErrInvalidEventRequest),
			want: true,
		},
		{
			name: "redis error",
			err:  errors.New("redis: connection refused"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPermanent(tt.err); got != tt.want {
				t.Errorf("isPermanent() = %v, want %v", got, tt.want)
			}
		})
	}
}
