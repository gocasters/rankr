package leaderboardscoring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapLeaderboardScoringToParam_ProjectID(t *testing.T) {
	projectID := "1001"
	result := LeaderboardQueryResult{LeaderboardRows: []LeaderboardEntry{
		{Rank: 1, UserID: "7", Score: 30},
		{Rank: 2, UserID: "8", Score: 10},
	}}

	tests := []struct {
		name      string
		projectID *string
	}{
		{name: "per-project rows carry the project", projectID: &projectID},
		{name: "global rows carry no project", projectID: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := mapLeaderboardScoringToParam(result, tt.projectID)

			assert.Equal(t, tt.projectID, res.ProjectID)
			assert.Len(t, res.LeaderboardRows, len(result.LeaderboardRows))
			for _, row := range res.LeaderboardRows {
				assert.Equal(t, tt.projectID, row.ProjectID)
			}
		})
	}
}
