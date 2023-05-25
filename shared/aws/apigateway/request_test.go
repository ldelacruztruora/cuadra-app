package apigateway

import (
	"testing"

	pemvalidation "bitbucket.org/truora/scrap-services/account/shared/pem-validation"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestParseRequestFormURLEncodedValid(t *testing.T) {
	c := require.New(t)

	values, err := ParseRequest(&events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body: "a=1&b=2&c=3&c=3",
	})

	c.Nil(err)
	c.NotNil(values)
	c.Equal("1", values["a"][0])
	c.Equal("2", values["b"][0])
	c.Equal("3", values["c"][0])
}

func TestParseRequestFormURLQueryValid(t *testing.T) {
	c := require.New(t)

	values, err := ParseRequest(&events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		MultiValueQueryStringParameters: map[string][]string{
			"a": {"1"},
			"b": {"2"},
			"c": {"3"},
		},
	})

	c.Nil(err)
	c.NotNil(values)
	c.Equal("1", values["a"][0])
	c.Equal("2", values["b"][0])
	c.Equal("3", values["c"][0])
}

func TestParseRequestJSONValid(t *testing.T) {
	c := require.New(t)

	values, err := ParseRequest(&events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"a": 1, "b": "2", "c": ["3", "3"]}`,
	})

	c.Nil(err)
	c.NotNil(values)
	c.Equal("1", values["a"][0])
	c.Equal("2", values["b"][0])
	c.Equal("3", values["c"][0])
}

func TestParseRequestJSONInvalid(t *testing.T) {
	c := require.New(t)

	values, err := ParseRequest(&events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `a=1`,
	})

	c.NotNil(err)
	c.Empty(values)
}

func TestParseJSON(t *testing.T) {
	c := require.New(t)

	values, err := ParseRequest(&events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{
			"arr": [1,2,"3"],
			"start_variables": {
				"welcome": "no",
				"int": 3,
				"arr2": [1,2,3],
				"hello": {
					"a": "something",
					"b": "else"
				}
			}
		}`,
	})

	c.NoError(err)

	t.Logf("values: %#v", values)

	c.Equal("no", values["start_variables.welcome"][0])
	c.Equal("3", values["start_variables.int"][0])
	c.Equal("something", values["start_variables.hello.a"][0])
	c.Equal("else", values["start_variables.hello.b"][0])
	c.EqualValues([]string{"1", "2", "3"}, values["arr"])
	c.EqualValues([]string{"1", "2", "3"}, values["start_variables.arr2"])
}

func TestParseRequestJWTValid(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	}

	values, err := ParseRequest(request)
	c.NoError(err)

	c.Equal("John Doe", values.Get("name"))
}

func TestParseRequestWithJWTInvalidPEM(t *testing.T) {
	c := require.New(t)

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): "dummyBadPEM",
				string(TenantIDKey):  "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.ErrorIs(err, pemvalidation.ErrInvalidPEM)
}

func TestParseRequestWithJWTInvalidOldPEM(t *testing.T) {
	c := require.New(t)

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): "dummyBadPEM",
				string(TenantIDKey):  "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.ErrorIs(err, pemvalidation.ErrInvalidPEM)
}

func TestParseRequestWithJWTOldKeyValid(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBkEuGeKjd4BiX84ITT85vZ
omm+2bHyBFlzWrZxdQIWxKcXKGBcC5SeP5otMrvz/P4K4V4d8ksm0WhtEoRhQbXR
bKxR/RNQWcZkbFR/uW/coNv0MO7kkysJhZ4MUNngdoSEoDsGQoMaFtAn2tnNksXj
3cpBZ5z4tSlkA9HG1ayh3sxsomn6u31wR7SkuMA9DxPzBDFnPzHvyenJs+uLnPjK
zK6jhvPQn2NnKnxUcbYk5lUdvU5trtYa+N7i8JDD098abfQS/nuGv6ouFpGpOX5v
FUxhqLu9VPEpBhBcHgYJwHi2SIFo2E3LJuNxoCyFGlUi3OglkafKZxWz/Ifl/YIv
AgMBAAE=
-----END PUBLIC KEY-----`

	oldPubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey):    pubPEM,
				string(OldJWTSecretKey): oldPubPEM,
				string(TenantIDKey):     "123",
			},
		},
	}

	values, err := ParseRequest(request)
	c.NoError(err)

	c.Equal("John Doe", values.Get("name"))
}

func TestParseRequestWithJWTOldKeyInvalid(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBkEuGeKjd4BiX84ITT85vZ
omm+2bHyBFlzWrZxdQIWxKcXKGBcC5SeP5otMrvz/P4K4V4d8ksm0WhtEoRhQbXR
bKxR/RNQWcZkbFR/uW/coNv0MO7kkysJhZ4MUNngdoSEoDsGQoMaFtAn2tnNksXj
3cpBZ5z4tSlkA9HG1ayh3sxsomn6u31wR7SkuMA9DxPzBDFnPzHvyenJs+uLnPjK
zK6jhvPQn2NnKnxUcbYk5lUdvU5trtYa+N7i8JDD098abfQS/nuGv6ouFpGpOX5v
FUxhqLu9VPEpBhBcHgYJwHi2SIFo2E3LJuNxoCyFGlUi3OglkafKZxWz/Ifl/YIv
AgMBAAE=
-----END PUBLIC KEY-----`

	oldPubPEM := "dummyBadPEM"

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey):    pubPEM,
				string(OldJWTSecretKey): oldPubPEM,
				string(TenantIDKey):     "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.ErrorIs(err, pemvalidation.ErrInvalidPEM)
}

func TestParseRequestWithJWTNotTenantPEMShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBkEuGeKjd4BiX84ITT85vZ
omm+2bHyBFlzWrZxdQIWxKcXKGBcC5SeP5otMrvz/P4K4V4d8ksm0WhtEoRhQbXR
bKxR/RNQWcZkbFR/uW/coNv0MO7kkysJhZ4MUNngdoSEoDsGQoMaFtAn2tnNksXj
3cpBZ5z4tSlkA9HG1ayh3sxsomn6u31wR7SkuMA9DxPzBDFnPzHvyenJs+uLnPjK
zK6jhvPQn2NnKnxUcbYk5lUdvU5trtYa+N7i8JDD098abfQS/nuGv6ouFpGpOX5v
FUxhqLu9VPEpBhBcHgYJwHi2SIFo2E3LJuNxoCyFGlUi3OglkafKZxWz/Ifl/YIv
AgMBAAE=
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjo0ODEzMjI4MTUzLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiMTIzIn0.6-kUiRGixp9NpMSC92Yq8BhFnhOQAAJcRdP4R22K9Dk8njaikF1ai75OAkORXSju2c9xB58oOtsYrjMFowWSIQQS2nlB0Yu7Ryl12dRLwbhmWNo5PnBGR2lAV4nSmE6SdAolzmSy3vj4HLiz-EHHcZBde7RHFRDd0Vsx518nApk1Q31vB2CQ51uFQoovfm2Il-7wd_yV5xwHP8g1kOP32sg7aWgY1wHEijjYViAofUKBG8ouA1Q0rAYjpB_O4qp3M1l-9sj8GGa3BOxpls4tLUPKUF9lke8oRT2Fm6pN2W606iX0u1P7WBcmAckEry4iaHTnZXSm-32RHYvAWyvykw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.Contains(err.Error(), "crypto/rsa: verification error")
}

func TestParseRequestWithJWTInvalidSigningFormatShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.Equal(err.Error(), errNoRSASigningMethod.Error())
}

func TestParseRequestWithJWTInvalidSigningFormatWithtOldKeyShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBITANBgkqhkiG9w0BAQEFAAOCAQ4AMIIBCQKCAQBkEuGeKjd4BiX84ITT85vZ
omm+2bHyBFlzWrZxdQIWxKcXKGBcC5SeP5otMrvz/P4K4V4d8ksm0WhtEoRhQbXR
bKxR/RNQWcZkbFR/uW/coNv0MO7kkysJhZ4MUNngdoSEoDsGQoMaFtAn2tnNksXj
3cpBZ5z4tSlkA9HG1ayh3sxsomn6u31wR7SkuMA9DxPzBDFnPzHvyenJs+uLnPjK
zK6jhvPQn2NnKnxUcbYk5lUdvU5trtYa+N7i8JDD098abfQS/nuGv6ouFpGpOX5v
FUxhqLu9VPEpBhBcHgYJwHi2SIFo2E3LJuNxoCyFGlUi3OglkafKZxWz/Ifl/YIv
AgMBAAE=
-----END PUBLIC KEY-----`

	oldPubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey):    pubPEM,
				string(OldJWTSecretKey): oldPubPEM,
				string(TenantIDKey):     "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.Equal(err.Error(), errNoRSASigningMethod.Error())
}

func TestParseRequestWithJWTInvalidIssuerShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoxNjY1MjYxODAyLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIiwiaXNzIjoiNDU2In0.eYTI2pkrZYzb1h7sB8vDzTz855G9YVaFZV4rICPesWPEdyaLxxRTwI5iOE848fA_c6oZU2k8WLIE23OKoDfdzoQ_Erv5DK9_7KoZJAvZVJe0waUPeFdZkR9wrTeAn59BEMa4TvrNIOZ4IsUKyS8e1-_durpyyVJgDPNsvJ4hbAQzQeo83zKrxuefMR6MEe3eARVNJmuRjl1js_w2GvQekOMvfxNYX5fdhxIvqkKV-pOeeD9GRShNTG2eAJfq-8u7NAX5GOWQ8NQy0bzE8ar0MARK-pfupxYeabnWcVH2RMurR5Qq6ZrdxkV27jgv3_kQV6jv8X2MfPd4xdvTzrSxAw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	} //iss = 456

	_, err := ParseRequest(request)
	c.Equal(err.Error(), errInvalidIssuer.Error())
}

func TestParseRequestWithJWTInvalidAudienceShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoxNjY1MjYxODAyLCJhdWQiOiJkdW1teSIsImlzcyI6IjEyMyJ9.OmBege_M53JxBkcd--FPCWF6caL8iUV2vm7RXi-rDceomgQYIJ0FCu8zjP6pN2nMQ0_zwWFEfUuRa5ZDuGY1zIeBvFOi1XT-_YCej4CGlXSLwsGq13aBXb_kV-qtK4bFbyiP0BPI3EDi8WElohWhckPOdPjQyo0MN8EEyXQykIr2gC_y_5pSfbdCJ9i-4HkGmDd9rhb8vq0CCyrgyUY8Y4LfVZZitIHGKn1QEfqGzjrzvFIAgXTHxhu8B8lbFa-PeqehBnkwRsECEUYdHERpZ3vdR1H1mRHjRUDmEFb_i7LkM-PbrrmV2-Vq5NueNpuWEg8M2-uZYkfIJv4F-M7sAw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	} //aud = dummy

	_, err := ParseRequest(request)
	c.Equal(err.Error(), errInvalidAudience.Error())
}

func TestParseRequestWithJWTJWTMissingIssuerShouldReturnError(t *testing.T) {
	c := require.New(t)

	pubPEM := `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA73qYTLEEFd3Gpsg6gHMx
X8t4n7cXzwWuOpcgXVt/ONL/FGQN1rGV1KKUxUZGBsDOy5G1acZFGRw3VO70vw4p
oR+AcN3a5J9UNchNAAcwyKxnMF5ALLwGfqyOLydb450HmO9vg5X2fyfIcrmcARjv
121I88bYsWxqejkQz72+mJ3e1Q75h7UNQGuQS6lXwBQ5ba6PGmWu2ylyPtO3HaZ2
oPYy0vNU0FKzLDYAfaCpxa3gtort29WYKZ70znhpZncNuy1JXWop0V3Dp/BH7uHd
oJZE2hzM1gbVENdIuaQi/Tf9QgHOhWSNbYrDyLeSHTkE5nP3uHOMdmXcK9R9Qs3p
LwIDAQAB
-----END PUBLIC KEY-----`

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoxNjY1MjYxODAyLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIn0.IHY_-mR5HEn3yrLQ458_kXxLwfDXJtGaA2m7eS8rm7FNq5NHZ7O6-iV1pq4noIYF_W5DurdRXsybiwpbkGQmXI414JBN258WbEW_RrXb66xMJh2X3ETAoIJIEN2dIQWoo1e5vNI2x4AiX7NGvNzSN0y8qE9gn9qEU6cnz94YxVgiMfN8P4aLoWviSzLh4W2Y4MxXE35YFMv_uvU-U8o9kEuEISqnhdnIMOMwGaGsbGU4VwQ6IAffCE8QUXa-JVMvcR9-z5Vf9NIv7JMlgppeyiqNX6ry0TVD_gHDamWwfsfVSwCQup_C2424QcgzJRvB7wcWyIPBLedQQHYpf6FdDw",
		RequestContext: events.APIGatewayProxyRequestContext{
			Authorizer: map[string]interface{}{
				string(JWTSecretKey): pubPEM,
				string(TenantIDKey):  "123",
			},
		},
	}

	_, err := ParseRequest(request)
	c.Equal(err.Error(), errInvalidIssuer.Error())
}

func TestParseRequestWithJWTMissingJWTSecret(t *testing.T) {
	c := require.New(t)

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
		Body: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZXhwIjoxNjY1MjYxODAyLCJhdWQiOiJWYWxpZGFjaW9uSWRlbnRpZGFkIn0.IHY_-mR5HEn3yrLQ458_kXxLwfDXJtGaA2m7eS8rm7FNq5NHZ7O6-iV1pq4noIYF_W5DurdRXsybiwpbkGQmXI414JBN258WbEW_RrXb66xMJh2X3ETAoIJIEN2dIQWoo1e5vNI2x4AiX7NGvNzSN0y8qE9gn9qEU6cnz94YxVgiMfN8P4aLoWviSzLh4W2Y4MxXE35YFMv_uvU-U8o9kEuEISqnhdnIMOMwGaGsbGU4VwQ6IAffCE8QUXa-JVMvcR9-z5Vf9NIv7JMlgppeyiqNX6ry0TVD_gHDamWwfsfVSwCQup_C2424QcgzJRvB7wcWyIPBLedQQHYpf6FdDw",
	}

	_, err := ParseRequest(request)
	c.ErrorIs(err, errJWTSecretNotFoundInRequestCtx)
}

func TestGetHeader(t *testing.T) {
	c := require.New(t)

	request := &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"content-type": "application/jwt",
		},
	}

	c.Equal("application/jwt", GetHeader(request, "Content-Type"))
	c.Equal("application/jwt", GetHeader(request, "content-type"))

	request = &events.APIGatewayProxyRequest{
		Headers: map[string]string{
			"Content-Type": "application/jwt",
		},
	}

	c.Equal("application/jwt", GetHeader(request, "Content-Type"))
	c.Equal("application/jwt", GetHeader(request, "content-type"))
}
