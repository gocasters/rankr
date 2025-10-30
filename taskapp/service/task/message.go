package task

import "errors"

var (
	ErrFailedToCreateTask      = errors.New("failed to create the task")
	ErrFailedToUpdateTask      = errors.New("failed to update task information")
	ErrFailedToBindData        = errors.New("failed to bind task data")
	ErrFailedToValidateRequest = errors.New("failed to validate request")
	ErrFailedToInsertTask      = errors.New("failed to insert task")
	ErrFailedToPublishEvent    = errors.New("failed to publish event in broker")
	ErrFailedToFindTask        = errors.New("failed to find task")
	ErrFailedToGetTask         = errors.New("failed to get task")
)

// Define constant messages
const (
	MessageValidationError   = "validation error"
	MessageUnexpectedError   = "unexpected error"
	MessageTaskCreated       = "task created successfully"
	MessageTaskUpdated       = "task updated successfully"
	MessageTaskNotFound      = "task not found"
	MessageInvalidTaskData   = "invalid task data"
	MessageDatabaseError     = "database error occurred"
	MessageRetryableError    = "retriable error, operation will be retried"
	MessageNonRetriableError = "non-retriable error, operation failed permanently"
)
