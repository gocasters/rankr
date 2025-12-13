package job

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileProcessor interface {
	Process(file *os.File) (ProcessResult, error)
}

type Repository interface {
	CreateJob(ctx context.Context, job Job) (uint, error)
	GetJobByIdempotencyKey(ctx context.Context, key string) (*Job, error)
	GetJobByFileHash(ctx context.Context, fileHash string) (*Job, error)
	GetJobByID(ctx context.Context, ID uint) (Job, error)
	UpdateStatus(ctx context.Context, jobID uint, status JobStatus) error
	GetFilePathByJobID(ctx context.Context, jobID uint) (string, error)
}

type Broker interface {
	Publish(ctx context.Context, pj ProduceJob) error
}

type FailRepository interface {
	InsertRecord(ctx context.Context, payload FailRecord) error
}

type FailJobRepository interface {
	Create(ctx context.Context, failRecord FailRecord) error
}

type ConfigJob struct {
	StoragePath string `koanf:"storage_path"`
	StreamKey   string `koanf:"stream_key"`
	CsvFile     string `koanf:"csv_file"`
	XlsxFile    string `koanf:"xlsx_file"`
	WorkerCount int    `koanf:"worker_count"`
	BufferSize  int    `koanf:"buffer_size"`
}

type Service struct {
	validation         Validate
	config             ConfigJob
	jobRepo            Repository
	broker             Broker
	contributorAdapter ContributorAdapter
	fileProcessor      FileProcessor
	failRepo           FailRepository
}

func NewService(validate Validate, cfg ConfigJob,
	repo Repository, broker Broker, contributorAdapter ContributorAdapter, failRecord FailRepository) Service {
	return Service{validation: validate, config: cfg, jobRepo: repo,
		broker: broker, contributorAdapter: contributorAdapter, failRepo: failRecord}
}

func (s Service) CreateJob(ctx context.Context, req ImportContributorRequest) (ImportContributorResponse, error) {
	if err := s.validation.validateFile(req); err != nil {
		return ImportContributorResponse{}, err
	}

	j, err := s.jobRepo.GetJobByIdempotencyKey(ctx, req.IdempotencyKey)
	if j != nil {
		return ImportContributorResponse{JobID: j.ID, Message: "The file with this idempotency key exists"}, nil
	}

	savePath := filepath.Join(s.config.StoragePath, fmt.Sprintf("%d_%s", time.Now().UnixNano(), req.FileName))
	dst, err := os.Create(savePath)
	if err != nil {
		return ImportContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed save file",
			Errors:          map[string]interface{}{"error": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	defer dst.Close()

	successSaveFile := false

	defer func() {
		if !successSaveFile {
			_ = os.Remove(savePath)
		}
	}()

	hashed := sha256.New()
	writer := io.MultiWriter(dst, hashed)
	if _, err := io.Copy(writer, req.File); err != nil {
		return ImportContributorResponse{}, err
	}
	fileHash := hex.EncodeToString(hashed.Sum(nil))

	existsJob, err := s.jobRepo.GetJobByFileHash(ctx, fileHash)
	if err != nil {
		return ImportContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed get file hash",
			Errors:          map[string]interface{}{"error db": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	if existsJob != nil {
		return ImportContributorResponse{JobID: existsJob.ID, Message: "file already exists"}, nil
	}

	job := Job{
		FileName:       req.FileName,
		FilePath:       savePath,
		IdempotencyKey: req.IdempotencyKey,
		Status:         Pending,
		CreatedAt:      time.Now(),
	}

	jobID, err := s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		logger.L().Error("failed create job", "error", err.Error())
		return ImportContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed to create job",
			Errors:          map[string]interface{}{"error db": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	job.ID = jobID

	if err := s.broker.Publish(ctx, ProduceJob{Key: s.config.StreamKey,
		JobID: job.ID, FilePath: job.FilePath}); err != nil {
		_ = s.jobRepo.UpdateStatus(ctx, job.ID, PendingToQueue)
		return ImportContributorResponse{
			JobID:   job.ID,
			Message: "job saved and will be queued shortly",
		}, nil
	}

	successSaveFile = true

	return ImportContributorResponse{JobID: job.ID, Message: "file received"}, nil
}

func (s Service) ProcessJob(ctx context.Context, jobID uint) (ProcessResult, error) {
	job, err := s.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		return ProcessResult{}, err
	}

	if job.Status != Pending {
		return ProcessResult{}, fmt.Errorf("invalid job status. job status is %s", job.Status)
	}

	filePath, err := s.jobRepo.GetFilePathByJobID(ctx, jobID)
	if err != nil {
		return ProcessResult{}, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var processor FileProcessor
	switch ext {
	case s.config.CsvFile:
		processor = CSVProcess{s.config.WorkerCount, s.config.BufferSize}
	case s.config.XlsxFile:
		processor = XLSXProcess{s.config.WorkerCount, s.config.BufferSize}
	default:
		return ProcessResult{}, fmt.Errorf("unsupported file extension: %s", ext)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return ProcessResult{}, err
	}
	defer file.Close()

	pResult, err := processor.Process(file)
	if err != nil {
		return ProcessResult{}, err
	}

	job.Status = Processing
	if err := s.jobRepo.UpdateStatus(ctx, jobID, Processing); err != nil {
		return ProcessResult{}, err
	}

	var successCount int
	var failCount int
	for _, record := range pResult.SuccessRecords {
		if err := s.contributorAdapter.UpsertContributor(ctx, record); err != nil {
			inErr := s.failRepo.InsertRecord(ctx, FailRecord{
				RecordNumber: record.RowNumber,
				Reason:       err.Error(),
				RawData:      record.mapToSlice(),
			})
			if inErr != nil {
				logger.L().Error("failed to insert fail record", "row", record.RowNumber, "error", inErr.Error())
			}

			failCount += 1

			continue
		}
		successCount++
	}

	for _, record := range pResult.FailRecords {
		if err := s.failRepo.InsertRecord(ctx, record); err != nil {
			logger.L().Error("failed fail record to repo",
				"row", record.RecordNumber, "error", err.Error())

			continue

		}

		failCount += 1
	}

	finalStatus := Failed
	if successCount > 0 && failCount > 0 {
		finalStatus = PartialSuccess
	} else if successCount > 0 {
		finalStatus = Success
	}

	if err := s.jobRepo.UpdateStatus(ctx, jobID, finalStatus); err != nil {
		logger.L().Error("failed update final job status", "error", err.Error())
		return ProcessResult{}, err
	}

	return ProcessResult{Total: pResult.Total, Success: successCount, Fail: failCount}, nil
}
