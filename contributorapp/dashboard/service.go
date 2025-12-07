package dashboard

import (
	"context"
	"fmt"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"mime/multipart"
	"strconv"
	"time"
)

const (
	CSV  string = "csv"
	XLSX string = "xlsx"
)

type FileProcessor interface {
	Process(file multipart.File, filename string) (ProcessResult, error)
}

type JobRepository interface {
	CreateJob(ctx context.Context, job Job) (uint, error)
	UpdateJob(ctx context.Context, job Job) error
}

type FailJobRepository interface {
	Create(ctx context.Context, failRecord FailRecord) error
}

type Service struct {
	validation     Validate
	jobRepo        JobRepository
	failJobRepo    FailJobRepository
	fileProcessor  FileProcessor
	contributorSvc ContributorAdapter
}

func NewService(validate Validate, repo JobRepository, contAdapter ContributorAdapter) Service {
	return Service{validation: validate, jobRepo: repo, contributorSvc: contAdapter}
}

func (s Service) ImportJob(ctx context.Context, req ImportJobRequest) (ImportJobResponse, error) {
	if err := s.validation.validateFile(req); err != nil {
		return ImportJobResponse{}, err
	}

	job := Job{
		FileName:  req.FileName,
		Status:    Pending,
		CreatedAt: time.Now(),
	}

	jobID, err := s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		logger.L().Error("failed create job", "error", err.Error())
		return ImportJobResponse{}, errmsg.ErrorResponse{
			Message:         "failed to create job",
			Errors:          map[string]interface{}{"error db": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	job.ID = jobID
	job.Status = Processing

	switch req.FileType {
	case XLSX:
		s.fileProcessor = XLSXProcess{}
	case CSV:
		s.fileProcessor = CSVProcess{}

	default:
		return ImportJobResponse{}, errmsg.ErrorResponse{
			Message:         "invalid input",
			Errors:          map[string]interface{}{"error": fmt.Sprintf("unsupported file extension %s: ", req.FileType)},
			InternalErrCode: statuscode.IntCodeInvalidParam,
		}
	}

	processFile, pErr := s.fileProcessor.Process(req.File, req.FileName)
	if pErr != nil {
		return ImportJobResponse{}, errmsg.ErrorResponse{
			Message:         "failed to process file",
			Errors:          map[string]interface{}{fmt.Sprintf("file %s", req.FileName): err.Error()},
			InternalErrCode: statuscode.IntCodeInvalidParam,
		}
	}

	if processFile.SuccessRecords != nil {
		for _, r := range processFile.SuccessRecords {
			err := s.contributorSvc.UpsertContributor(ctx, r)
			if err != nil {
				processFile.FailRecords = append(processFile.FailRecords, FailRecord{
					RecordNumber: r.RowNumber,
					Reason:       fmt.Sprintf("%v", err),
					RawData:      mapContributorRecordToSlice(r),
				})
			}
		}
	}

	if processFile.FailRecords != nil {
		for _, f := range processFile.FailRecords {
			err := s.failJobRepo.Create(ctx, f)
			// TODO search what to do the error
			logger.L().Error("error create fail job to database", "error", err)
		}
	}

	job.TotalRecords = processFile.Total
	job.SuccessCount = processFile.Success
	job.FailCount = processFile.Fail
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.UpdateJob(ctx, job); err != nil {
		logger.L().Error("failed to update job", "error", err.Error())
		// fmt.Errorf("failed to update job: %w", err)
		return ImportJobResponse{}, errmsg.ErrorResponse{
			Message:         "failed to update job",
			Errors:          map[string]interface{}{"repo_error": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	return ImportJobResponse{
		ID:           jobID,
		FileName:     req.FileName,
		TotalRecords: job.TotalRecords,
		SuccessCount: job.SuccessCount,
		FailCount:    job.FailCount,
		JobStatus:    job.Status,
	}, nil
}

func mapContributorRecordToSlice(cr ContributorRecord) []string {
	s := make([]string, 6)

	s[0] = strconv.Itoa(int(cr.GithubID))
	s[1] = cr.GithubUsername
	s[2] = cr.DisplayName
	s[3] = cr.ProfileImage
	s[4] = cr.Bio
	s[5] = cr.PrivacyMode

	return s
}
