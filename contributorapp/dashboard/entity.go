package dashboard

import (
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"time"
)

type Job struct {
	ID             uint
	TotalRecords   int
	SuccessCount   int
	FailCount      int
	IdempotencyKey string
	FileName       string
	FilePath       string
	Status         JobStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type IdempotencyRequest struct {
	ID             uint
	JobID          uint
	IdempotencyKey string
}

type FailJob struct {
	ID    uint
	JobID uint
	FailRecord
}

type JobStatus string

var (
	Pending    JobStatus = "pending"
	Success    JobStatus = "success"
	Failed     JobStatus = "failed"
	Processing JobStatus = "processing"
)

type ContributorRecord struct {
	RowNumber      int
	GithubID       int64
	GithubUsername string
	DisplayName    string
	ProfileImage   string
	Bio            string
	PrivacyMode    string
}

func (c ContributorRecord) mapContributorRecordToUpsertRequest() contributor.UpsertContributorRequest {
	return contributor.UpsertContributorRequest{
		GitHubID:       c.GithubID,
		GitHubUsername: c.GithubUsername,
		DisplayName:    c.DisplayName,
		ProfileImage:   c.ProfileImage,
		Bio:            c.Bio,
		PrivacyMode:    contributor.PrivacyMode(c.PrivacyMode),
	}
}

type FailRecord struct {
	RecordNumber int
	Reason       string
	RawData      []string
}

type ColumnName string

var (
	GithubID       ColumnName = "github_id"
	GithubUsername ColumnName = "github_username"
	DisplayName    ColumnName = "display_name"
	ProfileImage   ColumnName = "profile_image"
	Bio            ColumnName = "bio"
	PrivacyMode    ColumnName = "privacy_mode"
)

func (c ColumnName) String() string {
	switch c {
	case GithubID:
		return "github_id"
	case GithubUsername:
		return "github_username"
	case DisplayName:
		return "display_name"
	case ProfileImage:
		return "profile_image"
	case Bio:
		return "bio"
	case PrivacyMode:
		return "privacy_mode"
	}

	return ""
}
