package scheduler

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/logger"
)

// These tests pin the job names and defaults documented in
// docs/leaderboardstat-crons.md. If one fails, update the doc in the same PR.

func TestMain(m *testing.M) {
	// No FilePath: the log file goes next to the test binary, outside the repo.
	if err := logger.Init(logger.Config{Level: "error"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestNew_Defaults(t *testing.T) {
	tests := []struct {
		name string
		in   Config
		want Config
	}{
		{
			name: "empty config uses documented defaults",
			in:   Config{},
			want: Config{
				PublicLeaderboardCron:     "*/3 * * * *",
				DailyScoreCalculationCron: "0 2 * * *",
				JobContextTimeout:         3 * time.Minute,
			},
		},
		{
			name: "configured values are kept",
			in: Config{
				PublicLeaderboardCron:     "* * * * *",
				DailyScoreCalculationCron: "*/10 * * * *",
				JobContextTimeout:         time.Minute,
			},
			want: Config{
				PublicLeaderboardCron:     "* * * * *",
				DailyScoreCalculationCron: "*/10 * * * *",
				JobContextTimeout:         time.Minute,
			},
		},
		{
			name: "negative timeout falls back to default",
			in:   Config{JobContextTimeout: -time.Second},
			want: Config{
				PublicLeaderboardCron:     "*/3 * * * *",
				DailyScoreCalculationCron: "0 2 * * *",
				JobContextTimeout:         3 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(nil, tt.in)
			t.Cleanup(func() { _ = s.sch.Shutdown() })

			if s.cfg != tt.want {
				t.Errorf("cfg = %+v, want %+v", s.cfg, tt.want)
			}
		})
	}
}

func TestJobs_NamesAndTags(t *testing.T) {
	s := New(nil, Config{})
	t.Cleanup(func() { _ = s.sch.Shutdown() })

	ctx := context.Background()
	if err := s.dailyScoreCalculationJob(ctx); err != nil {
		t.Fatalf("dailyScoreCalculationJob: %v", err)
	}
	if err := s.publicLeaderboardJob(ctx); err != nil {
		t.Fatalf("publicLeaderboardJob: %v", err)
	}

	var names []string
	for _, j := range s.sch.Jobs() {
		names = append(names, j.Name())

		tags := j.Tags()
		if len(tags) != 1 || tags[0] != "leaderboardstat-service" {
			t.Errorf("job %q tags = %v, want [leaderboardstat-service]", j.Name(), tags)
		}
	}
	sort.Strings(names)

	want := []string{"daily-score-calculation", "public-leaderboard-update"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("job names = %v, want %v", names, want)
	}
}

func TestJobs_InvalidCronIsRejected(t *testing.T) {
	s := New(nil, Config{PublicLeaderboardCron: "not a cron"})
	t.Cleanup(func() { _ = s.sch.Shutdown() })

	if err := s.publicLeaderboardJob(context.Background()); err == nil {
		t.Error("publicLeaderboardJob with invalid crontab: want error, got nil")
	}
}
