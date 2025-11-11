package task

import (
	"context"
	"log/slog"
)

type Repository interface {
	CreateTask(ctx context.Context, param CreateTaskParam) error
	UpdateTaskByIssueNumber(ctx context.Context, param UpdateTaskParam) error
	GetTaskByIssueNumber(ctx context.Context, issueNumber int, repositoryName string) (*Task, error)
}

type Service struct {
	repository Repository
	validator  Validator
	logger     *slog.Logger
}

func NewService(
	repo Repository,
	validator Validator,
	logger *slog.Logger,
) Service {
	return Service{
		repository: repo,
		validator:  validator,
		logger:     logger,
	}
}

func (s Service) CreateTask(ctx context.Context, param CreateTaskParam) error {
	s.logger.Debug("Creating task from event",
		slog.Int("issue_number", param.IssueNumber),
		slog.String("title", param.Title),
	)

	return s.repository.CreateTask(ctx, param)
}

func (s Service) UpdateTask(ctx context.Context, param UpdateTaskParam) error {
	s.logger.Debug("Updating task from event",
		slog.Int("issue_number", param.IssueNumber),
		slog.String("state", param.State),
	)

	return s.repository.UpdateTaskByIssueNumber(ctx, param)
}
