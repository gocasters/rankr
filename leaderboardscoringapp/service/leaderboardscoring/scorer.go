package leaderboardscoring

import (
	"context"
	"slices"
	"strconv"
	"time"
)

type Award struct {
	UserID string
	Points int64
}

type Scorer interface {
	Score(ctx context.Context, req *EventRequest) ([]Award, error)
}

type legacyScorer struct{}

func NewLegacyScorer() Scorer {
	return legacyScorer{}
}

func (legacyScorer) Score(_ context.Context, req *EventRequest) ([]Award, error) {
	points := calculateScore(req.EventName)
	if points == 0 {
		return nil, nil
	}
	return []Award{{UserID: req.UserID, Points: points}}, nil
}

const (
	mergedPullRequestPoints = 10
	reviewPoints            = 3
	dailyReviewCap          = 5
)

type PendingReview struct {
	ReviewerID uint64
	ReviewedAt time.Time
}

type ReviewStore interface {
	AddPendingReview(ctx context.Context, prID uint64, review PendingReview) error
	TakePendingReviews(ctx context.Context, prID uint64) ([]PendingReview, error)
	IncrementDailyReviews(ctx context.Context, reviewerID uint64, day string) (int64, error)
}

type rulesV1Scorer struct {
	reviews  ReviewStore
	excluded map[uint64]struct{}
}

func NewRulesV1Scorer(reviews ReviewStore, excludedAccountIDs []uint64) Scorer {
	excluded := make(map[uint64]struct{}, len(excludedAccountIDs))
	for _, id := range excludedAccountIDs {
		excluded[id] = struct{}{}
	}
	return rulesV1Scorer{reviews: reviews, excluded: excluded}
}

func (s rulesV1Scorer) Score(ctx context.Context, req *EventRequest) ([]Award, error) {
	switch p := req.Payload.(type) {
	case PullRequestReviewPayload:
		return nil, s.holdReview(ctx, p, req.Timestamp)
	case PullRequestClosedPayload:
		return s.scoreClosedPullRequest(ctx, p)
	default:
		return nil, nil
	}
}

func (s rulesV1Scorer) holdReview(ctx context.Context, p PullRequestReviewPayload, reviewedAt time.Time) error {
	if p.State != ReviewStateApproved && p.State != ReviewStateChangesRequested {
		return nil
	}
	if p.ReviewerUserID == p.PrAuthorUserID || s.isExcluded(p.ReviewerUserID) {
		return nil
	}
	return s.reviews.AddPendingReview(ctx, p.PrID, PendingReview{ReviewerID: p.ReviewerUserID, ReviewedAt: reviewedAt})
}

func (s rulesV1Scorer) scoreClosedPullRequest(ctx context.Context, p PullRequestClosedPayload) ([]Award, error) {
	reviews, err := s.reviews.TakePendingReviews(ctx, p.PrID)
	if err != nil {
		return nil, err
	}
	if !p.Merged {
		return nil, nil
	}

	var awards []Award
	if !s.isExcluded(p.UserID) {
		awards = append(awards, Award{UserID: strconv.FormatUint(p.UserID, 10), Points: mergedPullRequestPoints})
	}

	slices.SortFunc(reviews, func(a, b PendingReview) int { return a.ReviewedAt.Compare(b.ReviewedAt) })
	for _, r := range reviews {
		count, err := s.reviews.IncrementDailyReviews(ctx, r.ReviewerID, r.ReviewedAt.UTC().Format(time.DateOnly))
		if err != nil {
			return nil, err
		}
		if count <= dailyReviewCap {
			awards = append(awards, Award{UserID: strconv.FormatUint(r.ReviewerID, 10), Points: reviewPoints})
		}
	}
	return awards, nil
}

func (s rulesV1Scorer) isExcluded(userID uint64) bool {
	_, ok := s.excluded[userID]
	return ok
}
