package job

import "mime/multipart"

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
