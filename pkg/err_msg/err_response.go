package errmsg

type ErrorResponse struct {
	Message         string                 `json:"message"`                       // General error message
	Errors          map[string]interface{} `json:"errors,omitempty"`              // Additional detail of error
	InternalErrCode string                 `json:"internal_error_code,omitempty"` // Custom error code (optional)
}

type ErrorType string

func (e ErrorResponse) Error() string {
	return e.Message
}

func NewError(err error, errorType ErrorType, message ...string) ErrorResponse {
	return ErrorResponse{}
}

func getMessage(err error, message []string) string {
	if len(message) > 0 {
		return message[0]
	}
	return err.Error()
}
