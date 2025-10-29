package task

import (
	"errors"
	"fmt"
)

var (
	ErrDatabaseConnectionFailed = errors.New("database connection failed")
	ErrDatabaseTimeout          = errors.New("database operation timeout")
	ErrTemporaryFailure         = errors.New("temporary failure, retry later")
)

var (
	ErrTaskNotFound          = errors.New("task not found")
	ErrTaskAlreadyExists     = errors.New("task already exists")
	ErrInvalidTaskData       = errors.New("invalid task data")
	ErrInvalidIssueNumber    = errors.New("invalid issue number")
	ErrInvalidRepositoryName = errors.New("invalid repository name")
	ErrInvalidState          = errors.New("invalid task state")
	ErrMissingRequiredField  = errors.New("missing required field")
	ErrConstraintViolation   = errors.New("database constraint violation")
	ErrForeignKeyViolation   = errors.New("foreign key constraint violation")
	ErrCheckConstraintFailed = errors.New("check constraint failed")
)

type TaskError struct {
	Err       error
	Retriable bool
	Message   string
}

func (e *TaskError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Err.Error()
}

func (e *TaskError) Unwrap() error {
	return e.Err
}

func (e *TaskError) IsRetriable() bool {
	return e.Retriable
}

func NewRetriableError(err error, message string) *TaskError {
	return &TaskError{
		Err:       err,
		Retriable: true,
		Message:   message,
	}
}

func NewNonRetriableError(err error, message string) *TaskError {
	return &TaskError{
		Err:       err,
		Retriable: false,
		Message:   message,
	}
}

func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	var taskErr *TaskError
	if errors.As(err, &taskErr) {
		return taskErr.IsRetriable()
	}

	if errors.Is(err, ErrDatabaseConnectionFailed) ||
		errors.Is(err, ErrDatabaseTimeout) ||
		errors.Is(err, ErrTemporaryFailure) {
		return true
	}

	if errors.Is(err, ErrTaskNotFound) ||
		errors.Is(err, ErrTaskAlreadyExists) ||
		errors.Is(err, ErrInvalidTaskData) ||
		errors.Is(err, ErrInvalidIssueNumber) ||
		errors.Is(err, ErrInvalidRepositoryName) ||
		errors.Is(err, ErrInvalidState) ||
		errors.Is(err, ErrMissingRequiredField) ||
		errors.Is(err, ErrConstraintViolation) ||
		errors.Is(err, ErrForeignKeyViolation) ||
		errors.Is(err, ErrCheckConstraintFailed) {
		return false
	}

	return true
}
