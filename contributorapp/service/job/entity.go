package job

import (
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"strconv"
	"time"
)

type Job struct {
	ID             uint
	TotalRecords   int
	SuccessCount   int
	FailCount      int
	IdempotencyKey string
	FileHash       string
	FileName       string
	FilePath       string
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProduceJob struct {
	Key      string
	JobID    uint
	FilePath string
}

type Status string

var (
	Pending        Status = "pending"
	PendingToQueue Status = "pending_to_queue"
	Success        Status = "success"
	Failed         Status = "failed"
	PartialSuccess Status = "partial_success"
	Processing     Status = "processing"
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

func (c ContributorRecord) mapToSlice() []string {
	s := make([]string, 0)
	idStr := strconv.Itoa(int(c.GithubID))
	s = append(s, idStr)
	s = append(s, c.GithubUsername)
	s = append(s, c.DisplayName)
	s = append(s, c.ProfileImage)
	s = append(s, c.Bio)
	s = append(s, string(c.PrivacyMode))

	return s
}

type FailRecord struct {
	ID           uint
	JobID        uint
	RecordNumber int
	Reason       string
	RawData      []string
	RetryCount   int
	LastError    *string
	CreatedAt    time.Time
	UpdatedAt    *time.Time
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
