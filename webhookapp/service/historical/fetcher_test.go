package historical

import (
	"testing"
)

func TestExtractReviewIDFromEventID(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		want      int64
		wantError bool
	}{
		{
			name:      "valid review event ID",
			eventID:   "historical-pr-42-review-7890",
			want:      7890,
			wantError: false,
		},
		{
			name:      "valid review event ID with large number",
			eventID:   "historical-pr-123-review-999999",
			want:      999999,
			wantError: false,
		},
		{
			name:      "invalid format - missing review keyword",
			eventID:   "historical-pr-42-7890",
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid format - too few parts",
			eventID:   "historical-pr-42",
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid format - not a review event",
			eventID:   "historical-pr-42-opened",
			want:      0,
			wantError: true,
		},
		{
			name:      "invalid format - non-numeric review ID",
			eventID:   "historical-pr-42-review-abc",
			want:      0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractReviewIDFromEventID(tt.eventID)
			if (err != nil) != tt.wantError {
				t.Errorf("extractReviewIDFromEventID() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("extractReviewIDFromEventID() = %v, want %v", got, tt.want)
			}
		})
	}
}
