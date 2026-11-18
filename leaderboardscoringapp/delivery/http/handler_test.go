package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/labstack/echo/v4"
)

type stubActivityStore struct {
	activity   []leaderboardscoring.ContributorActivity
	err        error
	gotUserID  string
	gotLimit   int
	calledOnce bool
}

func (s *stubActivityStore) AddProcessedScoreEvents(context.Context, []leaderboardscoring.ProcessedScoreEvent) error {
	return nil
}

func (s *stubActivityStore) GetVCSUserScoreBreakdown(context.Context, string) (leaderboardscoring.ScoreBreakdown, error) {
	return leaderboardscoring.ScoreBreakdown{}, nil
}

func (s *stubActivityStore) AddSnapshot(context.Context, []leaderboardscoring.SnapshotRow) error {
	return nil
}

func (s *stubActivityStore) GetContributorActivity(
	_ context.Context,
	userID string,
	limit int,
) ([]leaderboardscoring.ContributorActivity, error) {
	s.calledOnce = true
	s.gotUserID = userID
	s.gotLimit = limit

	if s.err != nil {
		return nil, s.err
	}

	return s.activity, nil
}

type activityResponseBody struct {
	UserID   string                                   `json:"user_id"`
	Limit    int                                      `json:"limit"`
	Activity []leaderboardscoring.ContributorActivity `json:"activity"`
}

func TestGetContributorActivity(t *testing.T) {
	occurredAt := time.Date(2026, time.November, 18, 10, 30, 0, 0, time.UTC)
	scored := []leaderboardscoring.ContributorActivity{{
		Type:           leaderboardscoring.PullRequestClosed,
		RepositoryName: "gocasters/rankr",
		SubjectType:    leaderboardscoring.SubjectTypePullRequest,
		SubjectNumber:  242,
		Points:         10,
		OccurredAt:     occurredAt,
	}}

	tests := []struct {
		name        string
		enabled     bool
		userID      string
		limit       string
		activity    []leaderboardscoring.ContributorActivity
		storeErr    error
		wantStatus  int
		wantLimit   int
		wantCalled  bool
		wantSubject string
	}{
		{
			name:       "disabled",
			enabled:    false,
			userID:     "90401",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "default limit",
			enabled:    true,
			userID:     "90401",
			activity:   scored,
			wantStatus: http.StatusOK,
			wantLimit:  20,
			wantCalled: true,
		},
		{
			name:       "explicit limit",
			enabled:    true,
			userID:     "90401",
			limit:      "5",
			activity:   scored,
			wantStatus: http.StatusOK,
			wantLimit:  5,
			wantCalled: true,
		},
		{
			name:       "limit is capped",
			enabled:    true,
			userID:     "90401",
			limit:      "500",
			activity:   scored,
			wantStatus: http.StatusOK,
			wantLimit:  50,
			wantCalled: true,
		},
		{
			name:       "invalid limit",
			enabled:    true,
			userID:     "90401",
			limit:      "0",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			enabled:    true,
			userID:     "90401",
			storeErr:   errors.New("store down"),
			wantStatus: http.StatusInternalServerError,
			wantLimit:  20,
			wantCalled: true,
		},
		{
			name:       "empty feed",
			enabled:    true,
			userID:     "90401",
			wantStatus: http.StatusOK,
			wantLimit:  20,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &stubActivityStore{activity: tt.activity, err: tt.storeErr}
			service := leaderboardscoring.NewService(
				store, nil, nil, "", leaderboardscoring.NewValidator(),
			)
			handler := NewHandler(service, tt.enabled)

			router := echo.New()
			router.GET("/v1/contributors/:user_id/activity", handler.GetContributorActivity)

			target := "/v1/contributors/" + tt.userID + "/activity"
			if tt.limit != "" {
				target += "?limit=" + tt.limit
			}

			request := httptest.NewRequest(http.MethodGet, target, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d; body: %s",
					tt.wantStatus, recorder.Code, recorder.Body.String())
			}

			if store.calledOnce != tt.wantCalled {
				t.Fatalf("expected store call %v, got %v", tt.wantCalled, store.calledOnce)
			}

			if !tt.wantCalled {
				return
			}

			if store.gotUserID != tt.userID {
				t.Errorf("expected user id %q, got %q", tt.userID, store.gotUserID)
			}

			if store.gotLimit != tt.wantLimit {
				t.Errorf("expected limit %d, got %d", tt.wantLimit, store.gotLimit)
			}

			if tt.wantStatus != http.StatusOK {
				return
			}

			var body activityResponseBody
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if body.UserID != tt.userID || body.Limit != tt.wantLimit {
				t.Errorf("unexpected response metadata: %+v", body)
			}

			if len(body.Activity) != len(tt.activity) {
				t.Fatalf("expected %d activity rows, got %d", len(tt.activity), len(body.Activity))
			}

			if len(body.Activity) > 0 {
				got := body.Activity[0]
				if got.Type != scored[0].Type ||
					got.RepositoryName != scored[0].RepositoryName ||
					got.SubjectType != scored[0].SubjectType ||
					got.SubjectNumber != scored[0].SubjectNumber ||
					got.Points != scored[0].Points ||
					!got.OccurredAt.Equal(occurredAt) {
					t.Errorf("unexpected activity row: %+v", got)
				}
			}
		})
	}
}

func TestGetContributorActivityRequiresUserID(t *testing.T) {
	router := echo.New()
	handler := NewHandler(
		leaderboardscoring.NewService(&stubActivityStore{}, nil, nil, "", leaderboardscoring.NewValidator()),
		true,
	)
	router.GET("/v1/contributors/:user_id/activity", handler.GetContributorActivity)

	request := httptest.NewRequest(http.MethodGet, "/v1/contributors/%20/activity", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "user_id is required") {
		t.Errorf("unexpected body: %s", recorder.Body.String())
	}
}
