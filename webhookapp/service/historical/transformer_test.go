package historical

import (
	"testing"
	"time"

	"github.com/gocasters/rankr/adapter/webhook/github"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
)

func TestTransformPRToEvents_OpenPR(t *testing.T) {
	now := time.Now()

	pr := &github.PullRequest{
		ID:     12345,
		Number: 42,
		State:  "open",
		Title:  "Test PR",
		User: github.User{
			ID:    100,
			Login: "testuser",
		},
		CreatedAt: now,
		Head: github.GitRef{
			Ref: "feature-branch",
			Repo: github.Repository{
				ID:       999,
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
		},
		Base: github.GitRef{
			Ref: "main",
			Repo: github.Repository{
				ID:       999,
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
		},
		Labels: []github.Label{
			{Name: "bug"},
			{Name: "urgent"},
		},
		Assignees: []github.User{
			{ID: 200},
			{ID: 300},
		},
	}

	events, err := TransformPRToEvents(pr)
	if err != nil {
		t.Fatalf("TransformPRToEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event (opened), got %d", len(events))
	}

	event := events[0]

	if event.EventName != eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED {
		t.Errorf("Expected PR_OPENED event, got %v", event.EventName)
	}

	if event.Id != "historical-pr-42-opened" {
		t.Errorf("Expected ID 'historical-pr-42-opened', got %s", event.Id)
	}

	if event.Provider != eventpb.EventProvider_EVENT_PROVIDER_GITHUB {
		t.Errorf("Expected GITHUB provider, got %v", event.Provider)
	}

	if event.RepositoryId != 999 {
		t.Errorf("Expected repo ID 999, got %d", event.RepositoryId)
	}

	payload := event.GetPrOpenedPayload()
	if payload == nil {
		t.Fatal("PR opened payload is nil")
	}

	if payload.UserId != 100 {
		t.Errorf("Expected user ID 100, got %d", payload.UserId)
	}

	if payload.PrNumber != 42 {
		t.Errorf("Expected PR number 42, got %d", payload.PrNumber)
	}

	if payload.Title != "Test PR" {
		t.Errorf("Expected title 'Test PR', got %s", payload.Title)
	}

	if payload.BranchName != "feature-branch" {
		t.Errorf("Expected branch 'feature-branch', got %s", payload.BranchName)
	}

	if payload.TargetBranch != "main" {
		t.Errorf("Expected target branch 'main', got %s", payload.TargetBranch)
	}

	if len(payload.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(payload.Labels))
	}

	if len(payload.Assignees) != 2 {
		t.Errorf("Expected 2 assignees, got %d", len(payload.Assignees))
	}
}

func TestTransformPRToEvents_ClosedMergedPR(t *testing.T) {
	now := time.Now()
	closedAt := now.Add(24 * time.Hour)

	mergerUser := &github.User{
		ID:    500,
		Login: "merger",
	}

	pr := &github.PullRequest{
		ID:       12345,
		Number:   42,
		State:    "closed",
		Title:    "Test PR",
		Merged:   true,
		ClosedAt: &closedAt,
		MergedBy: mergerUser,
		User: github.User{
			ID:    100,
			Login: "testuser",
		},
		CreatedAt:    now,
		Additions:    150,
		Deletions:    50,
		ChangedFiles: 10,
		Commits:      5,
		Head: github.GitRef{
			Ref: "feature-branch",
			Repo: github.Repository{
				ID:       999,
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
		},
		Base: github.GitRef{
			Ref: "main",
			Repo: github.Repository{
				ID:       999,
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
		},
		Labels:    []github.Label{},
		Assignees: []github.User{},
	}

	events, err := TransformPRToEvents(pr)
	if err != nil {
		t.Fatalf("TransformPRToEvents failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 events (opened + closed), got %d", len(events))
	}

	openedEvent := events[0]
	closedEvent := events[1]

	if openedEvent.EventName != eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED {
		t.Errorf("Expected first event to be PR_OPENED, got %v", openedEvent.EventName)
	}

	if closedEvent.EventName != eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED {
		t.Errorf("Expected second event to be PR_CLOSED, got %v", closedEvent.EventName)
	}

	if closedEvent.Id != "historical-pr-42-closed" {
		t.Errorf("Expected ID 'historical-pr-42-closed', got %s", closedEvent.Id)
	}

	payload := closedEvent.GetPrClosedPayload()
	if payload == nil {
		t.Fatal("PR closed payload is nil")
	}

	if payload.UserId != 100 {
		t.Errorf("Expected user ID 100, got %d", payload.UserId)
	}

	if payload.MergerUserId != 500 {
		t.Errorf("Expected merger ID 500, got %d", payload.MergerUserId)
	}

	if payload.CloseReason != eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED {
		t.Errorf("Expected MERGED close reason, got %v", payload.CloseReason)
	}

	if payload.Merged == nil || !*payload.Merged {
		t.Error("Expected Merged to be true")
	}

	if payload.Additions != 150 {
		t.Errorf("Expected 150 additions, got %d", payload.Additions)
	}

	if payload.Deletions != 50 {
		t.Errorf("Expected 50 deletions, got %d", payload.Deletions)
	}

	if payload.FilesChanged != 10 {
		t.Errorf("Expected 10 files changed, got %d", payload.FilesChanged)
	}

	if payload.CommitsCount != 5 {
		t.Errorf("Expected 5 commits, got %d", payload.CommitsCount)
	}
}

func TestTransformReviewToEvent(t *testing.T) {
	now := time.Now()

	pr := &github.PullRequest{
		ID:     12345,
		Number: 42,
		User: github.User{
			ID:    100,
			Login: "author",
		},
		Base: github.GitRef{
			Repo: github.Repository{
				ID:       999,
				Name:     "test-repo",
				FullName: "owner/test-repo",
			},
		},
	}

	review := &github.Review{
		ID: 7890,
		User: github.User{
			ID:    200,
			Login: "reviewer",
		},
		State:       "APPROVED",
		SubmittedAt: now,
	}

	event, err := TransformReviewToEvent(review, pr)
	if err != nil {
		t.Fatalf("TransformReviewToEvent failed: %v", err)
	}

	if event.EventName != eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED {
		t.Errorf("Expected REVIEW_SUBMITTED event, got %v", event.EventName)
	}

	if event.Id != "historical-pr-42-review-7890" {
		t.Errorf("Expected ID 'historical-pr-42-review-7890', got %s", event.Id)
	}

	payload := event.GetPrReviewPayload()
	if payload == nil {
		t.Fatal("PR review payload is nil")
	}

	if payload.ReviewerUserId != 200 {
		t.Errorf("Expected reviewer ID 200, got %d", payload.ReviewerUserId)
	}

	if payload.PrAuthorUserId != 100 {
		t.Errorf("Expected PR author ID 100, got %d", payload.PrAuthorUserId)
	}

	if payload.PrNumber != 42 {
		t.Errorf("Expected PR number 42, got %d", payload.PrNumber)
	}

	if payload.State != eventpb.ReviewState_REVIEW_STATE_APPROVED {
		t.Errorf("Expected APPROVED state, got %v", payload.State)
	}
}

func TestMapReviewState(t *testing.T) {
	tests := []struct {
		input    string
		expected eventpb.ReviewState
	}{
		{"APPROVED", eventpb.ReviewState_REVIEW_STATE_APPROVED},
		{"CHANGES_REQUESTED", eventpb.ReviewState_REVIEW_STATE_CHANGES_REQUESTED},
		{"COMMENTED", eventpb.ReviewState_REVIEW_STATE_COMMENTED},
		{"UNKNOWN", eventpb.ReviewState_REVIEW_STATE_UNSPECIFIED},
		{"", eventpb.ReviewState_REVIEW_STATE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := mapReviewState(tt.input)
			if result != tt.expected {
				t.Errorf("mapReviewState(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractLabels(t *testing.T) {
	labels := []github.Label{
		{Name: "bug"},
		{Name: "enhancement"},
		{Name: "urgent"},
	}

	result := extractLabels(labels)

	if len(result) != 3 {
		t.Fatalf("Expected 3 labels, got %d", len(result))
	}

	expected := []string{"bug", "enhancement", "urgent"}
	for i, label := range result {
		if label != expected[i] {
			t.Errorf("Label[%d] = %q, want %q", i, label, expected[i])
		}
	}
}

func TestExtractAssignees(t *testing.T) {
	assignees := []github.User{
		{ID: 100, Login: "user1"},
		{ID: 200, Login: "user2"},
	}

	result := extractAssignees(assignees)

	if len(result) != 2 {
		t.Fatalf("Expected 2 assignees, got %d", len(result))
	}

	expected := []uint64{100, 200}
	for i, id := range result {
		if id != expected[i] {
			t.Errorf("Assignee[%d] = %d, want %d", i, id, expected[i])
		}
	}
}
