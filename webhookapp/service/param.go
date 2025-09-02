package service

type IssueCommentCreatedRequest struct {
	Issue      Issue        `json:"issue"`
	Comment    IssueComment `json:"comment"`
	Repository Repository   `json:"repository"`
	Sender     User         `json:"sender"`
}
type IssueCommentCreatedResponse struct{}

type IssueCommentDeletedRequest struct {
	Issue      Issue        `json:"issue"`
	Comment    IssueComment `json:"comment"`
	Repository Repository   `json:"repository"`
	Sender     User         `json:"sender"`
}
type IssueCommentDeletedResponse struct{}

type IssueOpenedRequest struct {
	Issue      Issue      `json:"issue"`
	Type       IssueType  `json:"types"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}
type IssueOpenedResponse struct{}

type IssueClosedRequest struct {
	Issue      Issue      `json:"issue"`
	Type       IssueType  `json:"types"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}
type IssueClosedResponse struct{}

//types IssueTypedRequest struct {
//	Issue      Issue      `json:"issue"`
//	Type       IssueType  `json:"types"`
//	Label      *Label     `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//types IssueTypedResponse struct{}
//
//types IssueUntypedRequest struct {
//	Issue      Issue      `json:"issue"`
//	Type       IssueType  `json:"types"`
//	Label      *Label     `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//types IssueUntypedResponse struct{}
//
//types LabelCreatedRequest struct {
//	Label      Label      `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//types LabelCreatedResponse struct{}
//types LabelDeletedRequest struct {
//	Label      Label      `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//types LabelDeletedResponse struct{}

type PullRequestOpenedRequest struct {
	Number      int32       `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}
type PullRequestOpenedResponse struct{}

type PullRequestClosedRequest struct {
	Number      int32       `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}
type PullRequestClosedResponse struct{}

//types PullRequestReviewCommentCreatedRequest struct {
//	PullRequest PullRequest `json:"pull_request"`
//	Comment     interface{} `json:"comment"`
//	Repository  Repository  `json:"repository"`
//	Sender      User        `json:"sender"`
//}
//types PullRequestReviewCommentCreatedResponse struct{}
//
//types PullRequestReviewCommentDeletedRequest struct {
//	PullRequest PullRequest `json:"pull_request"`
//	Comment     interface{} `json:"comment"`
//	Repository  Repository  `json:"repository"`
//	Sender      User        `json:"sender"`
//}
//types PullRequestReviewCommentDeletedResponse struct{}

type PullRequestReviewSubmittedRequest struct {
	Review      PullRequestReview `json:"review"`
	PullRequest PullRequest       `json:"pull_request"`
	Repository  Repository        `json:"repository"`
	Sender      User              `json:"sender"`
}
type PullRequestReviewSubmittedResponse struct{}

type PushRequest struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
	HeadCommit *Commit    `json:"head_commit"`
	Commits    []Commit   `json:"commits"`
	Sender     User       `json:"sender"`
}
type PushResponse struct{}
