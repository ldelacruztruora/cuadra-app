package apigateway

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewError(t *testing.T) {
	c := require.New(t)

	err := NewError("my message", 230)
	response := NewErrorResponse(err)

	c.Equal(230, response.StatusCode)
	c.JSONEq(`{"code":10230,"http_code":230,"message":"my message"}`, response.Body)
}

func TestNewErrorResponseGenericError(t *testing.T) {
	c := require.New(t)

	response := NewErrorResponse(errors.New("dynamodb down"))
	c.Equal(500, response.StatusCode)
	c.JSONEq(`{"code":10500,"http_code":500,"message":"Internal server error, try again later"}`, response.Body)
}

func TestNewErrorResponseKnownError(t *testing.T) {
	c := require.New(t)

	response := NewErrorResponse(ErrNotFound)
	c.Equal(404, response.StatusCode)
	c.JSONEq(`{"code":10404,"http_code":404,"message":"Resource not found"}`, response.Body)
}

func TestErrorError(t *testing.T) {
	c := require.New(t)
	c.Equal("Resource not found (10404)", ErrNotFound.Error())
}

func TestNewInvalidRequestError(t *testing.T) {
	c := require.New(t)
	c.Equal("Invalid request: missing country (10400)", NewInvalidRequestError("missing country").Error())
}

func TestNewUnauthorizedRequestError(t *testing.T) {
	c := require.New(t)
	c.Equal("Unauthorized request: missing access to checks.users (10401)", NewUnauthorizedRequestError("missing access to checks.users").Error())
}

func TestNewForbiddenRequestError(t *testing.T) {
	c := require.New(t)
	c.Equal("Insufficient permissions: missing access to checks.users (10403)", NewForbiddenRequestError("missing access to checks.users").Error())
}

func TestNewNamedNotFoundError(t *testing.T) {
	c := require.New(t)
	c.Equal("check not found (10404)", NewNamedNotFoundError("check").Error())
}

func BenchmarkNewErrorResponse(b *testing.B) {
	c := require.New(b)

	for n := 0; n < b.N; n++ {
		response := NewErrorResponse(errors.New("dynamodb down"))
		c.Equal(http.StatusInternalServerError, response.StatusCode)
	}
}
