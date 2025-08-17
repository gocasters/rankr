package githubwebhook

import "time"

type Label struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
	Default bool   `json:"default"`
}
type Repository struct {
	ID       int64  `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"url"`
}

type User struct {
	AvatarURL string  `json:"avatar_url"`
	Email     *string `json:"email"`
	ID        int     `json:"id"`
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
	ID          int     `json:"id"`          // Unique identifier
	NodeID      string  `json:"node_id"`     // Node identifier
	Name        string  `json:"name"`        // Name of the issue type
	Description *string `json:"description"` // Description (nullable)
	Color       *string `json:"color"`       // Can be: gray, blue, green, yellow, orange, red, pink, purple, or null
	CreatedAt   *string `json:"created_at"`  // Creation timestamp (ISO 8601)
	UpdatedAt   *string `json:"updated_at"`  // Last update timestamp (ISO 8601)
	IsEnabled   bool    `json:"is_enabled"`  // Whether this type is enabled
}

type Issue struct {
	Assignee          *User      `json:"assignee"`
	Assignees         []*User    `json:"assignees"`
	AuthorAssociation string     `json:"author_association"`
	Body              *string    `json:"body"`
	ClosedAt          *time.Time `json:"closed_at"`
	Comments          int        `json:"comments"`
	CreatedAt         time.Time  `json:"created_at"`
	Draft             *bool      `json:"draft,omitempty"`
	ID                int        `json:"id"`
	Labels            []Label    `json:"labels"`
	Locked            bool       `json:"locked"`
	RepositoryURL     string     `json:"repository_url"`
	State             string     `json:"state"` // "open" or "closed"
	Title             string     `json:"title"`
	Type              *IssueType `json:"type"`
	UpdatedAt         time.Time  `json:"updated_at"`
	URL               string     `json:"url"`
	User              *User      `json:"user"`
}

type IssueComment struct {
	URL               string    `json:"url"`
	ID                int64     `json:"id"`
	NodeID            string    `json:"node_id"`
	User              User      `json:"user"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	AuthorAssociation string    `json:"author_association"`
	Body              string    `json:"body"`
}

type CommitComment struct {
	URL         string    `json:"url"`
	ID          int64     `json:"id"`
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
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	IssueURL string `json:"issue_url"`

	Title  string  `json:"title"`
	Body   *string `json:"body"`  // nullable
	State  string  `json:"state"` // open or closed
	Locked bool    `json:"locked"`
	Draft  bool    `json:"draft"`

	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	ClosedAt       *string `json:"closed_at"`        // nullable
	MergedAt       *string `json:"merged_at"`        // nullable
	MergeCommitSHA *string `json:"merge_commit_sha"` // nullable
	Merged         *bool   `json:"merged"`           // nullable
	Mergeable      *bool   `json:"mergeable"`        // nullable
	MergeableState string  `json:"mergeable_state"`
	Rebaseable     *bool   `json:"rebaseable"` // nullable

	Assignee           *User   `json:"assignee"`  // nullable
	Assignees          []*User `json:"assignees"` // nullable
	RequestedReviewers []*User `json:"requested_reviewers"`
	Labels             []Label `json:"labels"`
	MergedBy           *User   `json:"merged_by"` // nullable
	User               *User   `json:"user"`      // nullable

	Head GitRef `json:"head"`
	Base GitRef `json:"base"`

	ChangedFiles   int `json:"changed_files"`
	Additions      int `json:"additions"`
	Deletions      int `json:"deletions"`
	Comments       int `json:"comments"`
	Commits        int `json:"commits"`
	ReviewComments int `json:"review_comments"`

	ActiveLockReason *string    `json:"active_lock_reason"` // nullable: resolved, off-topic, too heated, spam
	AutoMerge        *AutoMerge `json:"auto_merge"`         // nullable

	Links PRLinks `json:"_links"`
}

type ReviewLinks struct {
	PullRequest Link `json:"pull_request"`
}
type PullRequestReview struct {
	ID                int64       `json:"id"`
	NodeID            string      `json:"node_id"`
	User              User        `json:"user"`
	Body              string      `json:"body"`
	CommitID          string      `json:"commit_id"`
	SubmittedAt       string      `json:"submitted_at"` // ISO8601 timestamp
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
	PullRequestReviewID int64              `json:"pull_request_review_id"`
	ID                  int64              `json:"id"`
	NodeID              string             `json:"node_id"`
	DiffHunk            string             `json:"diff_hunk"`
	Path                string             `json:"path"`
	CommitID            string             `json:"commit_id"`
	OriginalCommitID    string             `json:"original_commit_id"`
	User                User               `json:"user"`
	Body                string             `json:"body"`
	CreatedAt           string             `json:"created_at"` // ISO 8601 timestamp
	UpdatedAt           string             `json:"updated_at"` // ISO 8601 timestamp
	HtmlURL             string             `json:"html_url"`
	PullRequestURL      string             `json:"pull_request_url"`
	AuthorAssociation   string             `json:"author_association"`
	Links               ReviewCommentLinks `json:"_links"`
	StartLine           *int               `json:"start_line"`
	OriginalStartLine   *int               `json:"original_start_line"`
	StartSide           *string            `json:"start_side"`
	Line                *int               `json:"line"`
	OriginalLine        *int               `json:"original_line"`
	Side                string             `json:"side"` // RIGHT or LEFT
	OriginalPosition    int                `json:"original_position"`
	Position            int                `json:"position"`
	SubjectType         string             `json:"subject_type"` // "line"
}
