package task

import "errors"

var (
	ErrFailedToCreateTask          = errors.New("❌ failed to create the task")
	ErrFailedToUpdateTask          = errors.New("❌ failed to update Task's information")
	ErrFailedToBindData            = errors.New("❌ failed to bind task data")
	ErrFailedToValidateRequest     = errors.New("❌ failed to validate request")
	ErrFailedToInsertTask          = errors.New("❌ failed to insert task")
	ErrFailedToPublishEvent        = errors.New("❌ failed to publish event in broker")
	ErrFailedToFindTaskPhoneNumber = errors.New("❌ failed to find task by phone number")
	ErrFailedToLoginTask           = errors.New("❌ failed to login task ")
)

// Define constant messages
const (
	MessageValidationError = "validation error"
	MessageUnexpectedError = "unexpected error"
)
