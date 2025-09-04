package validator

import (
	"fmt"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	Flat   ErrorType = "flat"
	Nested ErrorType = "nested"
)

// ErrorType Type of showing errors
type ErrorType string

type Error struct {
	Fields map[string]interface{} `json:"fields,omitempty"`
	Err    string                 `json:"message"`
}

func (v Error) Error() string {
	return v.Err
}

func NewError(err error, errorType ErrorType, message ...string) Error {
	finalMessage := getMessage(err, message)

	fields := make(map[string]interface{})

	if errorsMap, ok := err.(validation.Errors); ok {
		if errorType == Flat {
			flattenValidationErrors("", errorsMap, fields)
		} else {
			fields = convertValidationErrorsToMap(errorsMap)
		}
	}

	return Error{
		Fields: fields,
		Err:    finalMessage,
	}
}

// Prepare final error message
func getMessage(err error, message []string) string {
	if len(message) > 0 {
		return message[0]
	}
	return err.Error()
}

func (v Error) StatusCode() int {
	if len(v.Fields) > 0 {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// show errors in flat shape
func flattenValidationErrors(prefix string, errorsMap validation.Errors, fields map[string]interface{}) {
	for field, validationErr := range errorsMap {
		fullFieldName := field
		if prefix != "" {
			fullFieldName = fmt.Sprintf("%s.%s", prefix, field)
		}

		if nestedErrors, ok := validationErr.(validation.Errors); ok {
			flattenValidationErrors(fullFieldName, nestedErrors, fields)
		} else {
			fields[fullFieldName] = validationErr.Error()
		}
	}
}

// show errors in nested shape
func convertValidationErrorsToMap(errorsMap validation.Errors) map[string]interface{} {
	result := make(map[string]interface{})
	for field, validationErr := range errorsMap {
		if nestedErrors, ok := validationErr.(validation.Errors); ok {
			result[field] = convertValidationErrorsToMap(nestedErrors)
		} else {
			result[field] = validationErr.Error()
		}
	}
	return result
}
