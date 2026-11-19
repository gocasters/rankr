package leaderboardscoring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	authorID uint64 = 2001
	botID    uint64 = 9001
)

func TestRulesV1Scorer(t *testing.T) {
	scorer := NewRulesV1Scorer([]uint64{botID})

	tests := []struct {
		name    string
		payload EventPayload
		want    int64
	}{
		{name: "merged pull request scores for the author", payload: PullRequestClosedPayload{UserID: authorID, Merged: true}, want: mergedPullRequestPoints},
		{name: "pull request closed without merge", payload: PullRequestClosedPayload{UserID: authorID, Merged: false}},
		{name: "merged pull request by an excluded account", payload: PullRequestClosedPayload{UserID: botID, Merged: true}},
		{name: "opened pull request", payload: PullRequestOpenedPayload{UserID: authorID}},
		{name: "review is not scored yet", payload: PullRequestReviewPayload{ReviewerUserID: authorID}},
		{name: "opened issue", payload: IssueOpenedPayload{UserID: authorID}},
		{name: "closed issue", payload: IssueClosedPayload{UserID: authorID, CloseReason: IssueCloseReasonCompleted}},
		{name: "issue comment", payload: IssueCommentedPayload{UserID: authorID}},
		{name: "push", payload: PushPayload{UserID: authorID}},
		{name: "missing payload", payload: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, scorer.Score(&EventRequest{Payload: tt.payload}))
		})
	}
}

func TestLegacyScorerMatchesCalculateScore(t *testing.T) {
	scorer := NewLegacyScorer()
	names := []EventName{PullRequestOpened, PullRequestClosed, PullRequestReview, IssueOpened, IssueClosed, IssueComment, CommitPush, EventName("unknown")}

	for _, name := range names {
		t.Run(string(name), func(t *testing.T) {
			require.Equal(t, calculateScore(string(name)), scorer.Score(&EventRequest{EventName: string(name)}))
		})
	}
}
