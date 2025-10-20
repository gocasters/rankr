package statuscode

import (
	"net/http"

	errmsg "github.com/gocasters/rankr/pkg/err_msg"
)

const (
	IntCodeInvalidParam   = "Invalid request parameter"
	IntCodeNotAuthorize   = "You need to authorize first"
	IntCodeNotPermission  = "You don't have permission"
	IntCodeRecordNotFound = "Record not found"
	IntCodeUnExpected     = "Unexpected issue"
	IntCodeNotFound       = "Not found"
)

// MapToHTTPStatusCode maps internal error codes to HTTP status codes
func MapToHTTPStatusCode(err errmsg.ErrorResponse) int {
	switch err.InternalErrCode {
	case IntCodeInvalidParam:
		return http.StatusBadRequest
	case IntCodeNotAuthorize:
		return http.StatusUnauthorized
	case IntCodeNotPermission:
		return http.StatusForbidden
	case IntCodeRecordNotFound:
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

// PostgreSQL Error Codes
const (
	ErrCodeUniqueViolation      = "23505" // Duplicate key
	ErrCodeForeignKeyViolation  = "23503" // Invalid FK
	ErrCodeSerializationFailure = "40001" // Transaction conflict
	ErrCodeDeadlockDetected     = "40P01" // Deadlock
	ErrCodeConnectionException  = "08000" // Connection problem
	ErrCodeConnectionNotExist   = "08003" // Connection closed
	ErrCodeConnectionFailure    = "08006" // Connection failed
	ErrTooManyConnections       = "53300" // Too many Connections
)
