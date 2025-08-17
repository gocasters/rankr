package githubwebhook

type EventType string
type PaLoadType string

const (
	EventTypeIssues            EventType = "issues"
	EventTypePullRequest       EventType = "pull_request"
	EventTypePullRequestReview EventType = "pull_request_review"

	TopicGithubUserActivity = "github.user.activity"

	PayloadTypeIssueOpened                PaLoadType = "issue_opened"
	PayloadTypeIssueClosed                PaLoadType = "issue_closed"
	PayloadTypePullRequestOpened          PaLoadType = "pull_request_opened"
	PayloadTypePullRequestClosed          PaLoadType = "pull_request_closed"
	PayloadTypePullRequestReviewSubmitted PaLoadType = "pull_request_review_submitted"
)

type ActivityEvent struct {
	Event       EventType   `json:"event"`
	Delivery    string      `json:"delivery"`
	PayloadType PaLoadType  `json:"payload_type"`
	Payload     interface{} `json:"body"`
}
