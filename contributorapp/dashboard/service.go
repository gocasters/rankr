package dashboard

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

type FileProcessor interface {
	Process(file multipart.File, filename string) (ProcessResult, error)
}

type JobRepository interface {
	CreateJob(ctx context.Context, job Job) (uint, error)
	GetJobByIdempotencyKey(ctx context.Context, key string) (*Job, error)
	GetJobByFileHash(ctx context.Context, fileHash string) (*Job, error)
}

type FailJobRepository interface {
	Create(ctx context.Context, failRecord FailRecord) error
}

type ConfigDashboard struct {
	StoragePath string `koanf:"storage_path"`
}

type Service struct {
	validation Validate
	config     ConfigDashboard
	jobRepo    JobRepository
}

func NewService(validate Validate, cfg ConfigDashboard, repo JobRepository) Service {
	return Service{validation: validate, config: cfg, jobRepo: repo}
}

func (s Service) ImportContributor(ctx context.Context, req ImportContributorRequest) (ImportContributorResponse, error) {
	if err := s.validation.validateFile(req); err != nil {
		return ImportContributorResponse{}, err
	}

	j, err := s.jobRepo.GetJobByIdempotencyKey(ctx, req.IdempotencyKey)
	if j != nil {
		return ImportContributorResponse{JobID: j.ID, Message: "The file with this idempotency key exists"}, nil
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, req.File); err != nil {
		return ImportContributorResponse{}, err
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	existsJob, err := s.jobRepo.GetJobByFileHash(ctx, fileHash)
	if err != nil {
		return ImportContributorResponse{}, err
	}

	if existsJob != nil {
		return ImportContributorResponse{JobID: existsJob.ID, Message: "file already exists"}, nil
	}

	savePath := filepath.Join(s.config.StoragePath, fmt.Sprintf("%d_%s", time.Now().UnixNano(), req.FileName))
	dst, err := os.Create(savePath)
	if err != nil {
		return ImportContributorResponse{}, err
	}

	defer dst.Close()

	req.File.Seek(0, 0)
	if _, err := io.Copy(dst, req.File); err != nil {
		return ImportContributorResponse{}, err
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

	return ImportContributorResponse{JobID: job.ID, Message: "file received"}, nil
}
