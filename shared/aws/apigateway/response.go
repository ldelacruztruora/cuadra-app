package apigateway

import (
	"encoding/json"
	"errors"
	"net/http"

	"bitbucket.org/truora/scrap-services/shared/env"
	"github.com/aws/aws-lambda-go/events"
)

var (
	origin = env.GetString("CORS_ORIGIN", "*")
)

// Response is the same as events.APIGatewayProxyResponse, left here for compatibility purposes
type Response = events.APIGatewayProxyResponse

// NewEmptyResponse creates a new response given a status code with empty body
func NewEmptyResponse(statusCode int) *events.APIGatewayProxyResponse {
	headers := map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": origin,
		"Cache-Control":               "no-store",
		"Pragma":                      "no-cache",
		"Strict-Transport-Security":   "max-age=63072000; includeSubdomains; preload",
	}

	if origin != "*" {
		headers["Access-Control-Allow-Credentials"] = "true"
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
	}
}

// NewJSONResponse creates a new JSON response given a serializable `v`
func NewJSONResponse(statusCode int, v interface{}) *events.APIGatewayProxyResponse {
	data, err := json.Marshal(v)
	if err != nil {
		return NewErrorResponse(errors.New("failed to marshal response"))
	}

	headers := map[string]string{
		"Content-Type":                "application/json",
		"Access-Control-Allow-Origin": origin,
		"Cache-Control":               "no-store",
		"Pragma":                      "no-cache",
		"Strict-Transport-Security":   "max-age=63072000; includeSubdomains; preload",
	}

	if origin != "*" {
		headers["Access-Control-Allow-Credentials"] = "true"
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(data),
		Headers:    headers,
	}
}

// NewRedirectionResponse creates a new response with Redirection
func NewRedirectionResponse(statusCode int, location string) *events.APIGatewayProxyResponse {
	headers := map[string]string{
		"Content-Type":                "text/htm",
		"Location":                    location,
		"Access-Control-Allow-Origin": origin,
		"Cache-Control":               "no-store",
		"Pragma":                      "no-cache",
		"Strict-Transport-Security":   "max-age=63072000; includeSubdomains; preload",
	}

	if origin != "*" {
		headers["Access-Control-Allow-Credentials"] = "true"
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
	}
}

// NewCSVResponse creates a new CSV response given a string `data`
func NewCSVResponse(statusCode int, data string) *events.APIGatewayProxyResponse {
	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       data,
		Headers: map[string]string{
			"Content-Type":              "text/csv",
			"Cache-Control":             "no-store",
			"Pragma":                    "no-cache",
			"Strict-Transport-Security": "max-age=63072000; includeSubdomains; preload",
		},
	}
}

// NewRedirectionResponseWithCookie creates a new response with Redirection and Cookie
func NewRedirectionResponseWithCookie(statusCode int, location string, cookie http.Cookie) *events.APIGatewayProxyResponse {
	headers := map[string]string{
		"Content-Type":                  "text/htm",
		"Location":                      location,
		"Access-Control-Allow-Origin":   origin,
		"Cache-Control":                 "no-store",
		"Pragma":                        "no-cache",
		"Strict-Transport-Security":     "max-age=63072000; includeSubdomains; preload",
		"Access-Control-Expose-Headers": "Set-Cookie",
		"Set-Cookie":                    cookie.String(),
	}

	if origin != "*" {
		headers["Access-Control-Allow-Credentials"] = "true"
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
	}
}

// NewJSONResponseWithCookie creates a new json response with a Cookie
func NewJSONResponseWithCookie(statusCode int, v interface{}, cookie http.Cookie, corsOrigin string) *events.APIGatewayProxyResponse {
	data, err := json.Marshal(v)
	if err != nil {
		return NewEmptyResponse(http.StatusInternalServerError)
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(data),
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      corsOrigin,
			"Access-Control-Allow-Credentials": "true",
			"Cache-Control":                    "no-store",
			"Pragma":                           "no-cache",
			"Strict-Transport-Security":        "max-age=63072000; includeSubdomains; preload",
			"Access-Control-Expose-Headers":    "Set-Cookie",
			"Set-Cookie":                       cookie.String(),
		},
	}
}
