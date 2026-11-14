package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
)

func TestRenderLeaderboard(t *testing.T) {
	tests := []struct {
		name string
		view leaderboardView
		want []string
	}{
		{
			name: "rows",
			view: leaderboardView{
				ProjectID: "5001",
				Rows: []leaderboardstat.PublicLeaderboardRow{
					{Rank: 1, UserID: 101, Score: 12},
				},
			},
			want: []string{
				`dir="rtl"`,
				"5001",
				"<td>1</td>",
				"<td>101</td>",
				"<td>12</td>",
			},
		},
		{
			name: "empty",
			view: leaderboardView{ProjectID: "5001"},
			want: []string{
				`dir="rtl"`,
				"هنوز امتیازی ثبت نشده است.",
			},
		},
		{
			name: "error",
			view: leaderboardView{ProjectID: "5001", Error: true},
			want: []string{
				`dir="rtl"`,
				"دریافت اطلاعات انجام نشد.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			if err := renderLeaderboard(&output, tt.view); err != nil {
				t.Fatalf("render leaderboard: %v", err)
			}

			for _, want := range tt.want {
				if !strings.Contains(output.String(), want) {
					t.Errorf("expected HTML containing %q, got: %s",
						want, output.String())
				}
			}
		})
	}
}
