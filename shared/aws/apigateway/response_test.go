package apigateway

import (
	"bytes"
	"net/http"
	"testing"

	"bitbucket.org/truora/scrap-services/shared/env"
	"bitbucket.org/truora/scrap-services/shared/safecsv"
	"github.com/stretchr/testify/require"
)

func TestNewJSONResponse(t *testing.T) {
	c := require.New(t)
	response := NewJSONResponse(http.StatusOK, map[string]string{"hello": "world"})

	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal("application/json", response.Headers["Content-Type"])
	c.JSONEq(`{"hello":"world"}`, response.Body)
}

func TestNewJSONResponseFromCustomOrigin(t *testing.T) {
	c := require.New(t)

	origin = "https://api.dummy.com"

	t.Cleanup(func() {
		origin = env.GetString("CORS_ORIGIN", "*")
	})

	response := NewJSONResponse(http.StatusOK, map[string]string{"hello": "world"})

	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal("application/json", response.Headers["Content-Type"])
	c.JSONEq(`{"hello":"world"}`, response.Body)
}

func TestNewEmptyResponse(t *testing.T) {
	c := require.New(t)
	response := NewEmptyResponse(http.StatusOK)

	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal("application/json", response.Headers["Content-Type"])
}

func TestNewEmptyResponseFromCustomOrigin(t *testing.T) {
	c := require.New(t)

	origin = "https://api.dummy.com"

	t.Cleanup(func() {
		origin = env.GetString("CORS_ORIGIN", "*")
	})

	response := NewEmptyResponse(http.StatusOK)

	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal("application/json", response.Headers["Content-Type"])
}

func TestNewRedirectionResponse(t *testing.T) {
	c := require.New(t)
	response := NewRedirectionResponse(http.StatusFound, "new url")

	c.Equal(http.StatusFound, response.StatusCode)
	c.Equal("text/htm", response.Headers["Content-Type"])
	c.Equal("new url", response.Headers["Location"])
}

func TestNewRedirectionResponseFromCustomOrigin(t *testing.T) {
	c := require.New(t)

	origin = "https://api.dummy.com"

	t.Cleanup(func() {
		origin = env.GetString("CORS_ORIGIN", "*")
	})

	response := NewRedirectionResponse(http.StatusFound, "new url")

	c.Equal(http.StatusFound, response.StatusCode)
	c.Equal("text/htm", response.Headers["Content-Type"])
	c.Equal("new url", response.Headers["Location"])
}

func TestNewCSVResponse(t *testing.T) {
	c := require.New(t)

	buffer := bytes.Buffer{}
	csvWriter := safecsv.NewWriter(',', &buffer)

	err := csvWriter.Write([]string{"Hello", "World"})
	c.NoError(err)

	csvWriter.Flush()

	response := NewCSVResponse(http.StatusOK, buffer.String())
	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal("text/csv", response.Headers["Content-Type"])
	c.Equal("Hello,World\n", response.Body)
}

func TestNewRedirectionResponseWithCookie(t *testing.T) {
	c := require.New(t)

	cookie := http.Cookie{
		Name:   "data",
		Value:  "cookie",
		Domain: "identity.truorastaging.com",
		Path:   "/",
		Secure: true,
		MaxAge: 300,
	}

	response := NewRedirectionResponseWithCookie(http.StatusFound, "new url", cookie)

	c.Equal(http.StatusFound, response.StatusCode)
	c.Equal("text/htm", response.Headers["Content-Type"])
	c.Equal("new url", response.Headers["Location"])
	c.Equal("Set-Cookie", response.Headers["Access-Control-Expose-Headers"])
	c.Equal("data=cookie; Path=/; Domain=identity.truorastaging.com; Max-Age=300; Secure", response.Headers["Set-Cookie"])
}

func TestNewRedirectionResponseWithCookieFromCustomOrigin(t *testing.T) {
	c := require.New(t)

	cookie := http.Cookie{
		Name:   "data",
		Value:  "cookie",
		Domain: "identity.truorastaging.com",
		Path:   "/",
		Secure: true,
		MaxAge: 300,
	}

	origin = "https://api.dummy.com"

	t.Cleanup(func() {
		origin = env.GetString("CORS_ORIGIN", "*")
	})

	response := NewRedirectionResponseWithCookie(http.StatusFound, "new url", cookie)

	c.Equal(http.StatusFound, response.StatusCode)
	c.Equal("text/htm", response.Headers["Content-Type"])
	c.Equal("new url", response.Headers["Location"])
	c.Equal("Set-Cookie", response.Headers["Access-Control-Expose-Headers"])
	c.Equal("data=cookie; Path=/; Domain=identity.truorastaging.com; Max-Age=300; Secure", response.Headers["Set-Cookie"])
}

func TestNewResponseWithCookie(t *testing.T) {
	c := require.New(t)

	cookie := http.Cookie{
		Name:  "data",
		Value: "cookie",
	}

	response := NewJSONResponseWithCookie(http.StatusOK, map[string]string{"hello": "world"}, cookie, "*")
	c.Equal(http.StatusOK, response.StatusCode)
	c.Equal(response.Headers["Set-Cookie"], "data=cookie")
}

func TestNewResponseWithCookieFail(t *testing.T) {
	c := require.New(t)

	cookie := http.Cookie{
		Name:  "data",
		Value: "cookie",
	}

	response := NewJSONResponseWithCookie(http.StatusOK, make(chan int), cookie, "*")
	c.Equal(http.StatusInternalServerError, response.StatusCode)
}

func BenchmarkNewJSONResponse(b *testing.B) {
	c := require.New(b)

	for n := 0; n < b.N; n++ {
		response := NewJSONResponse(http.StatusOK, map[string]string{"hello": "world"})
		c.Equal(http.StatusOK, response.StatusCode)
	}
}

func BenchmarkNewCSVResponse(b *testing.B) {
	c := require.New(b)

	buffer := bytes.Buffer{}
	csvWriter := safecsv.NewWriter(',', &buffer)

	err := csvWriter.Write([]string{"Hello", "World"})
	c.NoError(err)

	csvWriter.Flush()

	for n := 0; n < b.N; n++ {
		response := NewCSVResponse(http.StatusOK, buffer.String())
		c.Equal(http.StatusOK, response.StatusCode)
	}
}
