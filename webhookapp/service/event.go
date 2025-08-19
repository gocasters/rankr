package service

type EventType string
type PayLoadType string

const (
	EventTypeIssues            EventType = "issues"
	EventTypePullRequest       EventType = "pull_request"
	EventTypePullRequestReview EventType = "pull_request_review"

	TopicGithubUserActivity = "github.user.activity"

	PayloadTypeIssueOpened                PayLoadType = "issue_opened"
	PayloadTypeIssueClosed                PayLoadType = "issue_closed"
	PayloadTypePullRequestOpened          PayLoadType = "pull_request_opened"
	PayloadTypePullRequestClosed          PayLoadType = "pull_request_closed"
	PayloadTypePullRequestReviewSubmitted PayLoadType = "pull_request_review_submitted"
)

type ActivityEvent struct {
	Event       EventType   `json:"event"`
	Delivery    string      `json:"delivery"`
	PayloadType PayLoadType `json:"payload_type"`
	Payload     any         `json:"body"`
}
