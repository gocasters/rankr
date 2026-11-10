package rawevent

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
)

func TestHandleEventRejectsPayloadBeforeWrite(t *testing.T) {
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
			if err := (Handler{}).HandleEvent(msg); err == nil {
				t.Fatal("HandleEvent() = nil, want NACK before scoring write")
			}
		})
	}
}

func TestHandleProcessingResultOnlyAcknowledgesCompletedEvents(t *testing.T) {
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
		{name: "invalid event", err: leaderboardscoring.ErrInvalidEventRequest, want: leaderboardscoring.ErrInvalidEventRequest},
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
