package job

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrFailRecordNotFound = errors.New("fail record not found")
	ErrJobNotExists       = errors.New("job not exists")
)

type FileProcessor interface {
	Process(file *os.File) (ProcessResult, error)
}

type Repository interface {
	CreateJob(ctx context.Context, job Job) (uint, error)
	GetJobByIdempotencyKey(ctx context.Context, key string) (*Job, error)
	GetJobByFileHash(ctx context.Context, fileHash string) (*Job, error)
	GetJobByID(ctx context.Context, ID uint) (Job, error)
	UpdateStatus(ctx context.Context, jobID uint, status Status) error
	UpdateJob(ctx context.Context, job Job) error
}

type Broker interface {
	Publish(ctx context.Context, pj ProduceJob) error
}

type FailRepository interface {
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
	config             ConfigJob
	jobRepo            Repository
	broker             Broker
	contributorAdapter ContributorAdapter
	fileProcessor      FileProcessor
	failRepo           FailRepository
}

func NewService(cfg ConfigJob,
	repo Repository, broker Broker, contributorAdapter ContributorAdapter, failRecord FailRepository) Service {
	return Service{config: cfg, jobRepo: repo,
		broker: broker, contributorAdapter: contributorAdapter, failRepo: failRecord}
}

func (s Service) CreateImportJob(ctx context.Context, req contributor.ImportContributorRequest) (contributor.ImportContributorResponse, error) {

	j, err := s.jobRepo.GetJobByIdempotencyKey(ctx, req.IdempotencyKey)
	if j != nil {
		return contributor.ImportContributorResponse{JobID: j.ID, Message: "The file with this idempotency key exists"}, nil
	}

	if err != nil {
		if !errors.Is(err, ErrJobNotExists) {
			return contributor.ImportContributorResponse{}, errmsg.ErrorResponse{
				Message:         "failed to get job by idempotency key",
				Errors:          map[string]interface{}{"error": err.Error()},
				InternalErrCode: statuscode.IntCodeUnExpected,
			}
		}
	}

	savePath := filepath.Join(s.config.StoragePath, fmt.Sprintf("%d_%s", time.Now().UnixNano(), req.FileName))
	dst, err := os.Create(savePath)
	if err != nil {
		return contributor.ImportContributorResponse{}, errmsg.ErrorResponse{
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
		return contributor.ImportContributorResponse{}, err
	}
	fileHash := hex.EncodeToString(hashed.Sum(nil))

	existsJob, err := s.jobRepo.GetJobByFileHash(ctx, fileHash)
	if err != nil {
		if !errors.Is(err, ErrJobNotExists) {
			return contributor.ImportContributorResponse{}, errmsg.ErrorResponse{
				Message:         "failed get file hash",
				Errors:          map[string]interface{}{"error db": err.Error()},
				InternalErrCode: statuscode.IntCodeUnExpected,
			}
		}
	}

	if existsJob != nil {
		return contributor.ImportContributorResponse{JobID: existsJob.ID, Message: "file already exists"}, nil
	}

	job := Job{
		FileName:       req.FileName,
		FilePath:       savePath,
		IdempotencyKey: req.IdempotencyKey,
		FileHash:       fileHash,
		Status:         Pending,
		CreatedAt:      time.Now(),
	}

	jobID, err := s.jobRepo.CreateJob(ctx, job)
	if err != nil {
		logger.L().Error("failed create job", "error", err.Error())
		return contributor.ImportContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed to create job",
			Errors:          map[string]interface{}{"error db": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	job.ID = jobID

	if err := s.broker.Publish(ctx, ProduceJob{Key: s.config.StreamKey,
		JobID: job.ID, FilePath: job.FilePath}); err != nil {
		_ = s.jobRepo.UpdateStatus(ctx, job.ID, PendingToQueue)
		return contributor.ImportContributorResponse{
			JobID:   job.ID,
			Message: "job saved and will be queued shortly",
		}, nil
	}

	successSaveFile = true

	return contributor.ImportContributorResponse{JobID: job.ID, Message: "file received"}, nil
}

func (s Service) ProcessJob(ctx context.Context, jobID uint) (ProcessResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	job, err := s.jobRepo.GetJobByID(ctx, jobID)
	if err != nil {
		return ProcessResult{}, fmt.Errorf("failed to get job by id: %w", err)
	}

	if job.Status != Pending {
		return ProcessResult{}, fmt.Errorf("invalid job status: %s", job.Status)
	}

	job.Status = Processing
	if err := s.jobRepo.UpdateStatus(ctx, job.ID, Processing); err != nil {
		return ProcessResult{}, fmt.Errorf("update job status to processing: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(job.FilePath))
	var processor FileProcessor
	switch ext {
	case s.config.CsvFile:
		processor = CSVProcess{s.config.WorkerCount, s.config.BufferSize}
	case s.config.XlsxFile:
		processor = XLSXProcess{s.config.WorkerCount, s.config.BufferSize}
	default:
		job.Status = Failed
		_ = s.jobRepo.UpdateJob(ctx, job)
		return ProcessResult{}, fmt.Errorf("unsupported file extension: %s", ext)
	}

	file, err := os.Open(job.FilePath)
	if err != nil {
		job.Status = Failed
		_ = s.jobRepo.UpdateJob(ctx, job)
		return ProcessResult{}, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	pResult, err := processor.Process(file)
	if err != nil {
		job.Status = Failed
		_ = s.jobRepo.UpdateJob(ctx, job)
		return ProcessResult{}, fmt.Errorf("process file: %w", err)
	}

	successCount, failCount := 0, 0

	for _, record := range pResult.SuccessRecords {
		if err := s.contributorAdapter.UpsertContributor(ctx, record); err != nil {
			cErr := s.failRepo.Create(ctx, FailRecord{
				JobID:        jobID,
				RecordNumber: record.RowNumber,
				Reason:       err.Error(),
				RawData:      record.mapToSlice(),
				RetryCount:   0,
			})
			if cErr != nil {
				logger.L().Error("failed to insert fail record", "row", record.RowNumber, "error", cErr.Error())
			}
			failCount++
			continue
		}
		successCount++
	}

	for _, record := range pResult.FailRecords {
		record.JobID = jobID
		record.RetryCount = 0
		if err := s.failRepo.Create(ctx, record); err != nil {
			logger.L().Error("failed to insert fail record", "row", record.RecordNumber, "error", err.Error())
		}
		failCount++
	}

	finalStatus := Failed
	if successCount > 0 && failCount > 0 {
		finalStatus = PartialSuccess
	} else if successCount > 0 {
		finalStatus = Success
	}

	job.Status = finalStatus
	job.SuccessCount = successCount
	job.FailCount = failCount
	job.TotalRecords = pResult.Total
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.UpdateJob(ctx, job); err != nil {
		logger.L().Error("failed to update final job", "jobID", jobID, "error", err.Error())
		return ProcessResult{}, fmt.Errorf("update final job: %w", err)
	}

	return ProcessResult{Total: pResult.Total, Success: successCount, Fail: failCount}, nil
}
