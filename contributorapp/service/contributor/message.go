package contributor

import "errors"

var (
	ErrFailedToCreateContributor          = errors.New("❌ failed to create the contributor")
	ErrFailedToUpdateContributor          = errors.New("❌ failed to update Contributor's information")
	ErrFailedToBindData                   = errors.New("❌ failed to bind contributor data")
	ErrFailedToValidateRequest            = errors.New("❌ failed to validate request")
	ErrFailedToInsertContributor          = errors.New("❌ failed to insert contributor")
	ErrFailedToPublishEvent               = errors.New("❌ failed to publish event in broker")
	ErrFailedToFindContributorPhoneNumber = errors.New("❌ failed to find contributor by phone number")
	ErrFailedToLoginContributor           = errors.New("❌ failed to login contributor ")
)

// Define constant messages
const (
	MessageValidationError = "validation error"
	MessageUnexpectedError = "unexpected error"
)
