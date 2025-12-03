package dashboard

import "mime/multipart"

type ImportJobRequest struct {
	File     multipart.File `json:"file"`
	FileName string         `json:"file_name"`
	FileType string         `json:"file_type"`
}

type ImportJobResponse struct {
	ID           uint      `json:"id"`
	FileName     string    `json:"file_name"`
	TotalRecords int       `json:"total_records"`
	SuccessCount int       `json:"success_count"`
	FailCount    int       `json:"fail_count"`
	JobStatus    JobStatus `json:"job_status"`
}
