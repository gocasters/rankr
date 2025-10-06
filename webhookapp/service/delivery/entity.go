package delivery

import (
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"time"
)

type Label struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Default bool   `json:"default"`
}
type Repository struct {
	ID       uint64 `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"url"`
}

type User struct {
	AvatarURL string  `json:"avatar_url"`
	Email     *string `json:"email"`
	ID        uint64  `json:"id"`
	Login     string  `json:"login"`
	Name      *string `json:"name"`
	NodeID    string  `json:"node_id"`
}

type GitRef struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	SHA   string     `json:"sha"`
	User  User       `json:"user"`
	Repo  Repository `json:"repo"` // Define Repo struct as needed
}

type IssueType struct {
	ID          uint64     `json:"id"`          // Unique identifier
	NodeID      string     `json:"node_id"`     // Node identifier
	Name        string     `json:"name"`        // Name of the issue types
	Description *string    `json:"description"` // Description (nullable)
	Color       *string    `json:"color"`       // Can be: gray, blue, green, yellow, orange, red, pink, purple, or null
	CreatedAt   *time.Time `json:"created_at"`  // Creation timestamp (ISO 8601)
	UpdatedAt   *time.Time `json:"updated_at"`  // Last update timestamp (ISO 8601)
	IsEnabled   bool       `json:"is_enabled"`  // Whether this types is enabled
}

type Issue struct {
	Assignee          *User      `json:"assignee"`
	Assignees         []*User    `json:"assignees"`
	AuthorAssociation string     `json:"author_association"`
	Body              *string    `json:"body"`
	ClosedAt          *time.Time `json:"closed_at"`
	Comments          int32      `json:"comments"`
	CreatedAt         time.Time  `json:"created_at"`
	// Draft is PR-only; omit on Issue
	//Draft             *bool        `json:"draft,omitempty"`
	ID            uint64       `json:"id"`
	Number        int32        `json:"number"`
	Labels        []Label      `json:"labels"`
	Locked        bool         `json:"locked"`
	RepositoryURL string       `json:"repository_url"`
	State         string       `json:"state"` // "open" or "closed"
	StateReason   *string      `json:"state_reason"`
	Title         string       `json:"title"`
	Type          *IssueType   `json:"types"`
	UpdatedAt     time.Time    `json:"updated_at"`
	URL           string       `json:"url"`
	User          *User        `json:"user"`
	PullRequest   *PullRequest `json:"pull_request,omitempty"`
}

type IssueComment struct {
	URL               string    `json:"url"`
	ID                uint64    `json:"id"`
	NodeID            string    `json:"node_id"`
	User              User      `json:"user"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	AuthorAssociation string    `json:"author_association"`
	Body              string    `json:"body"`
}

type CommitComment struct {
	URL         string    `json:"url"`
	ID          uint64    `json:"id"`
	NodeID      string    `json:"node_id"`
	User        User      `json:"user"`
	CommitID    string    `json:"commit_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AuthorAssoc string    `json:"author_association"` //CONTRIBUTOR
	Body        string    `json:"body"`
}

type AutoMerge struct {
	MergeMethod string `json:"merge_method"`
	CommitTitle string `json:"commit_title"`
	CommitBody  string `json:"commit_message"`
	EnabledBy   User   `json:"enabled_by"`
}

type Link struct {
	Href string `json:"href"`
}

type PRLinks struct {
	Self     Link `json:"self"`
	HTML     Link `json:"html"`
	Issue    Link `json:"issue"`
	Comments Link `json:"comments"`
	Review   Link `json:"review_comments"`
	Review2  Link `json:"review_comment"`
	Commits  Link `json:"commits"`
	Statuses Link `json:"statuses"`
}

type PullRequest struct {
	URL      string `json:"url"`
	ID       uint64 `json:"id"`
	Number   int32  `json:"number"`
	NodeID   string `json:"node_id"`
	IssueURL string `json:"issue_url"`

	Title  string  `json:"title"`
	Body   *string `json:"body"`  // nullable
	State  string  `json:"state"` // open or closed
	Locked bool    `json:"locked"`
	Draft  bool    `json:"draft"`

	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ClosedAt       *time.Time `json:"closed_at"`        // nullable
	MergedAt       *time.Time `json:"merged_at"`        // nullable
	MergeCommitSHA *string    `json:"merge_commit_sha"` // nullable
	Merged         *bool      `json:"merged"`           // nullable
	Mergeable      *bool      `json:"mergeable"`        // nullable
	MergeableState string     `json:"mergeable_state"`
	Rebaseable     *bool      `json:"rebaseable"` // nullable

	Assignee           *User   `json:"assignee"`  // nullable
	Assignees          []*User `json:"assignees"` // nullable
	RequestedReviewers []*User `json:"requested_reviewers"`
	Labels             []Label `json:"labels"`
	MergedBy           *User   `json:"merged_by"` // nullable
	User               *User   `json:"user"`      // nullable

	Head GitRef `json:"head"`
	Base GitRef `json:"base"`

	ChangedFiles   int32 `json:"changed_files"`
	Additions      int32 `json:"additions"`
	Deletions      int32 `json:"deletions"`
	Comments       int32 `json:"comments"`
	Commits        int32 `json:"commits"`
	ReviewComments int32 `json:"review_comments"`

	ActiveLockReason *string    `json:"active_lock_reason"` // nullable: resolved, off-topic, too heated, spam
	AutoMerge        *AutoMerge `json:"auto_merge"`         // nullable

	Links PRLinks `json:"_links"`
}

type ReviewLinks struct {
	PullRequest Link `json:"pull_request"`
}
type PullRequestReview struct {
	ID                uint64      `json:"id"`
	NodeID            string      `json:"node_id"`
	User              User        `json:"user"`
	Body              string      `json:"body"`
	CommitID          string      `json:"commit_id"`
	SubmittedAt       time.Time   `json:"submitted_at"` // ISO8601 timestamp
	State             string      `json:"state"`        // e.g., "commented", "approved", etc.
	HtmlURL           string      `json:"html_url"`
	PullRequestURL    string      `json:"pull_request_url"`
	AuthorAssociation string      `json:"author_association"` // e.g., CONTRIBUTOR, MEMBER, etc.
	Links             ReviewLinks `json:"_links"`
}

type ReviewCommentLinks struct {
	PullRequest Link `json:"pull_request"`
}

type PullRequestReviewComment struct {
	URL                 string             `json:"url"`
	PullRequestReviewID uint64             `json:"pull_request_review_id"`
	ID                  uint64             `json:"id"`
	NodeID              string             `json:"node_id"`
	DiffHunk            string             `json:"diff_hunk"`
	Path                string             `json:"path"`
	CommitID            string             `json:"commit_id"`
	OriginalCommitID    string             `json:"original_commit_id"`
	User                User               `json:"user"`
	Body                string             `json:"body"`
	CreatedAt           time.Time          `json:"created_at"` // ISO 8601 timestamp
	UpdatedAt           time.Time          `json:"updated_at"` // ISO 8601 timestamp
	HtmlURL             string             `json:"html_url"`
	PullRequestURL      string             `json:"pull_request_url"`
	AuthorAssociation   string             `json:"author_association"`
	Links               ReviewCommentLinks `json:"_links"`
	StartLine           *int32             `json:"start_line"`
	OriginalStartLine   *int32             `json:"original_start_line"`
	StartSide           *string            `json:"start_side"`
	Line                *int32             `json:"line"`
	OriginalLine        *int32             `json:"original_line"`
	Side                string             `json:"side"` // RIGHT or LEFT
	OriginalPosition    int32              `json:"original_position"`
	Position            int32              `json:"position"`
	SubjectType         string             `json:"subject_type"` // "line"
}

type Commit struct {
	ID        string   `json:"id"`
	TreeID    string   `json:"tree_id"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	URL       string   `json:"url"`
	Author    Author   `json:"author"`
	Committer Author   `json:"committer"`
	Added     []string `json:"added"`
	Removed   []string `json:"removed"`
	Modified  []string `json:"modified"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
type Pusher struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type EventType string

const (
	EventTypeIssues            EventType = "issues"
	EventTypeIssueComment      EventType = "issue_comment"
	EventTypePullRequest       EventType = "pull_request"
	EventTypePullRequestReview EventType = "pull_request_review"
	EventTypePush              EventType = "push"
)

type Topic string

const (
	TopicGithubIssues       Topic = "github.issues"
	TopicGithubIssueComment Topic = "github.issue_comments"
	TopicGithubPullRequest  Topic = "github.pull_requests"
	TopicGithubReview       Topic = "github.reviews"
	TopicGithubPush         Topic = "github.pushes"
)

// //////////////////////////////////
type GitHubDelivery struct {
	ID             int64     `json:"id"`
	GUID           string    `json:"guid"`
	DeliveredAt    time.Time `json:"delivered_at"`
	Redelivery     bool      `json:"redelivery"`
	Duration       float64   `json:"duration"`
	Status         string    `json:"status"`
	StatusCode     int       `json:"status_code"`
	Event          string    `json:"event"`
	Action         string    `json:"action,omitempty"`
	InstallationID *int64    `json:"installation_id,omitempty"`
	RepositoryID   *int64    `json:"repository_id,omitempty"`
}

type DeliveryComparison struct {
	MissingDeliveries []GitHubDelivery
	FailedDeliveries  []GitHubDelivery
	ExtraEvents       []*eventpb.Event
}
