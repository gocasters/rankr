package rawevent

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
)

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
