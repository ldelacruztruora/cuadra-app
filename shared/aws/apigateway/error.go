package apigateway

import (
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// Error represents an API error
type Error struct {
	Code     int    `json:"code"`
	HTTPCode int    `json:"http_code"`
	Message  string `json:"message"`
}

// Error returns the error message
func (e *Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.Code)
}

// NewError method to initialize custom error
func NewError(message string, code int) error {
	return &Error{
		Code:     10000 + code,
		HTTPCode: code,
		Message:  message,
	}
}

// NewErrorResponse returns an error response
func NewErrorResponse(err error) *events.APIGatewayProxyResponse {
	var knownError *Error
	if errors.As(err, &knownError) {
		return NewJSONResponse(knownError.HTTPCode, knownError)
	}

	return NewJSONResponse(ErrInternalError.HTTPCode, ErrInternalError)
}

// NewInvalidRequestError returns an error for invalid requests that can be rendered in a HTTP response
func NewInvalidRequestError(message string) error {
	return &Error{
		Code:     10400,
		HTTPCode: 400,
		Message:  "Invalid request: " + message,
	}
}

// NewUnauthorizedRequestError returns an error for invalid requests that can be rendered in a HTTP response
func NewUnauthorizedRequestError(message string) error {
	return &Error{
		Code:     10401,
		HTTPCode: 401,
		Message:  "Unauthorized request: " + message,
	}
}

// NewForbiddenRequestError returns an error for forbidden requests that can be rendered in a HTTP response
func NewForbiddenRequestError(message string) error {
	return &Error{
		Code:     10403,
		HTTPCode: 403,
		Message:  "Insufficient permissions: " + message,
	}
}

// NewNamedNotFoundError returns a not found error specifying the name of the not found resource
func NewNamedNotFoundError(resourceName string) error {
	return &Error{
		Code:     10404,
		HTTPCode: 404,
		Message:  resourceName + " not found",
	}
}

// List of all known errors
var (
	// ErrInternalError is returned when there's an internal error that must be retried
	ErrInternalError = &Error{
		Code:     10500,
		HTTPCode: 500,
		Message:  "Internal server error, try again later",
	}

	// ErrInvalidRequest is returned when the client request is invalid
	ErrInvalidRequest = &Error{
		Code:     10400,
		HTTPCode: 400,
		Message:  "Invalid request",
	}

	// ErrMissingLanguage is returned when the details translator fails
	ErrMissingLanguage = &Error{
		Code:     10400,
		HTTPCode: 400,
		Message:  "Translator Failed",
	}

	// ErrResourceNotFound is returned when the requested resource is not found
	ErrNotFound = &Error{
		Code:     10404,
		HTTPCode: 404,
		Message:  "Resource not found",
	}

	// ErrUnauthorized is returned when the requested resource is unauthorized
	ErrUnauthorized = &Error{
		Code:     10401,
		HTTPCode: 401,
		Message:  "Unauthorized request",
	}

	// ErrForbidden is returned when the user doens't have enough permissions
	ErrForbidden = &Error{
		Code:     10403,
		HTTPCode: 403,
		Message:  "Insufficient permissions",
	}
)
