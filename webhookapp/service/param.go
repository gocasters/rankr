package service

type CommitCommentCreatedRequest struct {
	Comment    CommitComment `json:"comment"`
	Repository Repository    `json:"repository"`
	Sender     User          `json:"sender"`
}
type CommitCommentCreatedResponse struct{}

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
	Type       IssueType  `json:"type"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}
type IssueOpenedResponse struct{}

type IssueClosedRequest struct {
	Issue      Issue      `json:"issue"`
	Type       IssueType  `json:"type"`
	Repository Repository `json:"repository"`
	Sender     User       `json:"sender"`
}
type IssueClosedResponse struct{}

//type IssueTypedRequest struct {
//	Issue      Issue      `json:"issue"`
//	Type       IssueType  `json:"type"`
//	Label      *Label     `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//type IssueTypedResponse struct{}
//
//type IssueUntypedRequest struct {
//	Issue      Issue      `json:"issue"`
//	Type       IssueType  `json:"type"`
//	Label      *Label     `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//type IssueUntypedResponse struct{}
//
//type LabelCreatedRequest struct {
//	Label      Label      `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//type LabelCreatedResponse struct{}
//type LabelDeletedRequest struct {
//	Label      Label      `json:"label"`
//	Repository Repository `json:"repository"`
//	Sender     User       `json:"sender"`
//}
//type LabelDeletedResponse struct{}

type PullRequestOpenedRequest struct {
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}
type PullRequestOpenedResponse struct{}

type PullRequestClosedRequest struct {
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      User        `json:"sender"`
}
type PullRequestClosedResponse struct{}

//type PullRequestReviewCommentCreatedRequest struct {
//	PullRequest PullRequest `json:"pull_request"`
//	Comment     interface{} `json:"comment"`
//	Repository  Repository  `json:"repository"`
//	Sender      User        `json:"sender"`
//}
//type PullRequestReviewCommentCreatedResponse struct{}
//
//type PullRequestReviewCommentDeletedRequest struct {
//	PullRequest PullRequest `json:"pull_request"`
//	Comment     interface{} `json:"comment"`
//	Repository  Repository  `json:"repository"`
//	Sender      User        `json:"sender"`
//}
//type PullRequestReviewCommentDeletedResponse struct{}

type PullRequestReviewSubmittedRequest struct {
	Review      PullRequestReview `json:"review"`
	PullRequest PullRequest       `json:"pull_request"`
	Repository  Repository        `json:"repository"`
	Sender      User              `json:"sender"`
}
type PullRequestReviewSubmittedResponse struct{}
