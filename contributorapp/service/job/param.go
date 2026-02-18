package job

import (
	"mime/multipart"
	"time"
)

type ImportContributorRequest struct {
	File           multipart.File `json:"file"`
	FileName       string         `json:"file_name"`
	FileType       string         `json:"file_type"`
	IdempotencyKey string         `json:"idempotency_key"`
}

type ImportContributorResponse struct {
	JobID   uint   `json:"job_id"`
	Message string `json:"message"`
}

type GetJobResponse struct {
	ID             uint      `json:"ID"`
	TotalRecords   int       `json:"total_records,omitempty"`
	SuccessCount   int       `json:"success_count,omitempty"`
	FailCount      int       `json:"fail_count,omitempty"`
	IdempotencyKey string    `json:"idempotency_key"`
	FileHash       string    `json:"file_hash"`
	FileName       string    `json:"file_name"`
	FilePath       string    `json:"file_path"`
	Status         Status    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

type GetFailRecordsResponse struct {
	ID           uint          `json:"id"`
	JobID        uint          `json:"job_id"`
	RecordNumber int           `json:"record_number"`
	Reason       string        `json:"reason"`
	RawData      []string      `json:"raw_data"`
	RetryCount   int           `json:"retry_count"`
	LastError    string        `json:"last_error"`
	ErrType      RecordErrType `json:"err_type"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    *time.Time    `json:"updated_at"`
}
