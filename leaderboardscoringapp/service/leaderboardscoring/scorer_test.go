package leaderboardscoring

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	authorID   uint64 = 2001
	reviewerID uint64 = 2002
	botID      uint64 = 9001
	prID       uint64 = 4000001
)

var errStoreDown = errors.New("store down")

type memoryReviewStore struct {
	pending map[uint64][]PendingReview
	daily   map[string]int64
	err     error
}

func (m *memoryReviewStore) AddPendingReview(_ context.Context, prID uint64, review PendingReview) error {
	reviews := slices.DeleteFunc(m.pending[prID], func(r PendingReview) bool { return r.ReviewerID == review.ReviewerID })
	m.pending[prID] = append(reviews, review)
	return m.err
}

func (m *memoryReviewStore) TakePendingReviews(_ context.Context, prID uint64) ([]PendingReview, error) {
	reviews := m.pending[prID]
	delete(m.pending, prID)
	return reviews, m.err
}

func (m *memoryReviewStore) IncrementDailyReviews(_ context.Context, reviewerID uint64, day string) (int64, error) {
	m.daily[day+":"+strconv.FormatUint(reviewerID, 10)]++
	return m.daily[day+":"+strconv.FormatUint(reviewerID, 10)], nil
}

func newScorer(err error) Scorer {
	return NewRulesV1Scorer(&memoryReviewStore{pending: map[uint64][]PendingReview{}, daily: map[string]int64{}, err: err}, []uint64{botID})
}

func review(reviewer, pr uint64, state ReviewState) *EventRequest {
	at := time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC)
	return &EventRequest{Timestamp: at, Payload: PullRequestReviewPayload{ReviewerUserID: reviewer, PrAuthorUserID: authorID, PrID: pr, State: state}}
}

func closed(pr, author uint64, merged bool) *EventRequest {
	return &EventRequest{Payload: PullRequestClosedPayload{UserID: author, PrID: pr, Merged: merged}}
}

func scoreAll(t *testing.T, scorer Scorer, reqs ...*EventRequest) []Award {
	t.Helper()
	var awards []Award
	for _, req := range reqs {
		got, err := scorer.Score(context.Background(), req)
		require.NoError(t, err)
		awards = append(awards, got...)
	}
	return awards
}

func TestRulesV1Scorer(t *testing.T) {
	author := Award{UserID: "2001", Points: mergedPullRequestPoints}
	reviewer := Award{UserID: "2002", Points: reviewPoints}
	merged := closed(prID, authorID, true)

	tests := []struct {
		name   string
		events []*EventRequest
		want   []Award
	}{
		{"merged pull request scores for the author", []*EventRequest{merged}, []Award{author}},
		{"pull request closed without merge", []*EventRequest{closed(prID, authorID, false)}, nil},
		{"merged pull request by an excluded account", []*EventRequest{closed(prID, botID, true)}, nil},
		{"approved review scores when the pull request merges", []*EventRequest{review(reviewerID, prID, ReviewStateApproved), merged}, []Award{author, reviewer}},
		{"changes requested review scores when the pull request merges", []*EventRequest{review(reviewerID, prID, ReviewStateChangesRequested), merged}, []Award{author, reviewer}},
		{"review alone does not score", []*EventRequest{review(reviewerID, prID, ReviewStateApproved)}, nil},
		{"commented review", []*EventRequest{review(reviewerID, prID, ReviewStateCommented), merged}, []Award{author}},
		{"self review", []*EventRequest{review(authorID, prID, ReviewStateApproved), merged}, []Award{author}},
		{"review by an excluded account", []*EventRequest{review(botID, prID, ReviewStateApproved), merged}, []Award{author}},
		{"review on a pull request closed without merge", []*EventRequest{review(reviewerID, prID, ReviewStateApproved), closed(prID, authorID, false), merged}, []Award{author}},
		{"one review per reviewer per pull request", []*EventRequest{review(reviewerID, prID, ReviewStateChangesRequested), review(reviewerID, prID, ReviewStateApproved), merged}, []Award{author, reviewer}},
		{"opened pull request", []*EventRequest{{Payload: PullRequestOpenedPayload{UserID: authorID}}}, nil},
		{"closed issue", []*EventRequest{{Payload: IssueClosedPayload{UserID: authorID, CloseReason: IssueCloseReasonCompleted}}}, nil},
		{"issue comment", []*EventRequest{{Payload: IssueCommentedPayload{UserID: authorID}}}, nil},
		{"push", []*EventRequest{{Payload: PushPayload{UserID: authorID}}}, nil},
		{"missing payload", []*EventRequest{{}}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, scoreAll(t, newScorer(nil), tt.events...))
		})
	}
}

func TestRulesV1ScorerDailyReviewCap(t *testing.T) {
	scorer := newScorer(nil)
	reviewed := 0
	for pr := uint64(1); pr <= dailyReviewCap+1; pr++ {
		awards := scoreAll(t, scorer, review(reviewerID, pr, ReviewStateApproved), closed(pr, authorID, true))
		reviewed += len(awards) - 1
	}
	require.Equal(t, dailyReviewCap, reviewed)
}

func TestRulesV1ScorerStoreError(t *testing.T) {
	scorer := newScorer(errStoreDown)
	for _, req := range []*EventRequest{review(reviewerID, prID, ReviewStateApproved), closed(prID, authorID, true)} {
		awards, err := scorer.Score(context.Background(), req)
		require.ErrorIs(t, err, errStoreDown)
		require.Empty(t, awards)
	}
}
