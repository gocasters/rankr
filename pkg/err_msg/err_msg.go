package errmsg

import "errors"

var (
	ErrValidationFailed     = errors.New("input validation failed")
	ErrUnexpectedError      = errors.New("unexpected error occurred")
	ErrInvalidRequestFormat = errors.New("invalid request format")
	ErrGetUserInfo          = errors.New("get user info failed")
	ErrFailedDecodeBase64   = errors.New("decode data to base 64 failed")
	ErrFailedUnmarshalJson  = errors.New("unmarshal data to JSON failed")
	ErrUnauthorized         = errors.New("unauthorized")
)

// Define constant messages generally
const (
	MessageMissingXUserData  = "Missing X-User-Info header"
	MessageInvalidBase64     = "Invalid Base64 data"
	MessageInvalidJsonFormat = "Invalid JSON format"
	ServerError              = "Internal server error"
)
