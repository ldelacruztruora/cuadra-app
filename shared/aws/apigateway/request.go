// Package apigateway contains scraper microservices and account services
package apigateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	pemvalidation "bitbucket.org/truora/scrap-services/account/shared/pem-validation"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v4"
)

var (
	errNoRSASigningMethod            = NewInvalidRequestError("no rsa signing method")
	errNoClaimsFound                 = NewInvalidRequestError("no claims found")
	audience                         = "ValidacionIdentidad"
	errInvalidIssuer                 = NewInvalidRequestError("invalid issuer found")
	errInvalidAudience               = NewInvalidRequestError("invalid audience found")
	errJWTSecretNotFoundInRequestCtx = errors.New("jwt secret not found in request context")
	errTenantIDNotFoundInCtx         = errors.New("tenant ID not found in request context")
)

// Request is a request from the AWS API Gateway when using the default Lambda proxy.
type Request = events.APIGatewayProxyRequest

// ContextKey type used to represent the key of a context value
type ContextKey string

const (
	// JWTSecretKey key value for getting the tenant JWTSecret from context
	JWTSecretKey ContextKey = "jwt_secret"
	// OldJWTSecretKey key value for getting the tenant OldJWTSecret from context
	OldJWTSecretKey ContextKey = "old_jwt_secret" // #nosec This is not a hardcoded credential
	// TenantIDKey key value for getting the tenant ID from context
	TenantIDKey ContextKey = "tenant_id"

	normalizedHeadersKey = "request-normalized-headers"
)

func claimsValue(claims jwt.MapClaims, key string) string {
	value, ok := claims[key].(string)
	if !ok || value == "" {
		return ""
	}

	return value
}

// GetHeader gets value for the specified header
// Request headers are case insensitive by specification https://www.rfc-editor.org/rfc/rfc7230#section-3.2
func GetHeader(req *events.APIGatewayProxyRequest, header string) string {
	normalizeRequestHeaders(req)

	return req.Headers[strings.ToLower(header)]
}

func normalizeRequestHeaders(req *events.APIGatewayProxyRequest) {
	if req.Headers[normalizedHeadersKey] == "true" {
		return
	}

	normalizedHeaders := make(map[string]string, len(req.Headers))

	for k, v := range req.Headers {
		normalizedHeaders[strings.ToLower(k)] = v
	}

	normalizedHeaders[normalizedHeadersKey] = "true"
	req.Headers = normalizedHeaders
}

// ParseRequest parses the request body depending on the content type
func ParseRequest(req *events.APIGatewayProxyRequest) (url.Values, error) {
	contentType := GetHeader(req, "Content-Type")

	if contentType == "application/json" {
		return parseJSONBody(req.Body)
	}

	if contentType == "application/jwt" {
		return parseJWTBody(req.RequestContext.Authorizer, req.Body)
	}

	if req.Body != "" {
		return url.ParseQuery(req.Body)
	}

	return req.MultiValueQueryStringParameters, nil
}

func parseJSONBody(body string) (url.Values, error) {
	parsedJSONBody := map[string]interface{}{}

	err := json.Unmarshal([]byte(body), &parsedJSONBody)
	if err != nil {
		return url.Values{}, err
	}

	return interfaceAsValues("", parsedJSONBody), nil
}

func getCtxJWTValues(context map[string]interface{}) (string, string, string, error) {
	JWTSecret, ok := context[string(JWTSecretKey)].(string)
	if !ok {
		return "", "", "", errJWTSecretNotFoundInRequestCtx
	}

	OldJWTSecret, ok := context[string(OldJWTSecretKey)].(string)
	if !ok {
		OldJWTSecret = ""
	}

	tenantID, ok := context[string(TenantIDKey)].(string)
	if !ok {
		return "", "", "", errTenantIDNotFoundInCtx
	}

	return JWTSecret, OldJWTSecret, tenantID, nil
}

func getPayloadFromJWT(JWTSecret string, body string, tenantID string) (string, error) {
	pubKey, err := pemvalidation.ParsePEMPK(JWTSecret)
	if err != nil {
		return "", err
	}

	claims, err := parseJWT(body, pubKey, tenantID)
	if err != nil {
		return "", err
	}

	jsonBody, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	return string(jsonBody), nil
}

func parseJWTBody(context map[string]interface{}, body string) (url.Values, error) {
	JWTSecret, OldJWTSecret, tenantID, err := getCtxJWTValues(context)
	if err != nil {
		return nil, err
	}

	jsonBody, err := getPayloadFromJWT(JWTSecret, body, tenantID)
	if err == nil {
		return parseJSONBody(jsonBody)
	}

	if OldJWTSecret == "" {
		return nil, err
	}

	jsonBody, err = getPayloadFromJWT(OldJWTSecret, body, tenantID)
	if err != nil {
		return nil, err
	}

	return parseJSONBody(jsonBody)
}

func checkTokenJWT(token *jwt.Token, audience string, issuer string) error {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return errNoRSASigningMethod
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errNoClaimsFound
	}

	iss := claimsValue(claims, "iss")
	if iss != issuer {
		return errInvalidIssuer
	}

	aud := claimsValue(claims, "aud")
	if aud != audience {
		return errInvalidAudience
	}

	return nil
}

func parseJWT(content string, publicKey interface{}, issuer string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(content, func(token *jwt.Token) (interface{}, error) {
		err := checkTokenJWT(token, audience, issuer)
		if err != nil {
			return "", err
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errNoClaimsFound
	}

	return claims, nil
}

func interfaceAsValues(key string, i interface{}) url.Values {
	switch v := i.(type) {
	case string:
		return url.Values{key: {v}}
	case []interface{}:
		return sliceToValues(key, v)
	case map[string]interface{}:
		return mapToValues(key, v)
	default:
		return url.Values{key: {fmt.Sprintf("%v", v)}}
	}
}

func sliceToValues(key string, v []interface{}) url.Values {
	values := url.Values{}

	for _, innerValue := range v {
		subValues := interfaceAsValues(key, innerValue)

		for k, v := range subValues {
			values[k] = append(values[k], v...)
		}
	}

	return values
}

func mapToValues(key string, v map[string]interface{}) url.Values {
	values := url.Values{}

	for k, v := range v {
		if key != "" {
			k = key + "." + k
		}

		subValues := interfaceAsValues(k, v)

		for kk, vv := range subValues {
			values[kk] = vv
		}
	}

	return values
}
