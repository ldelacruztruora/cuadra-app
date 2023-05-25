package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"bitbucket.org/truora/scrap-services/logger"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetInflightKeys(b *testing.B) {
	params := url.Values{"a": {"b"}}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}

	for i := 0; i < b.N; i++ {
		_, err := Get(context.Background(), "http://httpbin.org/get", headers, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestTimeoutDuration(t *testing.T) {
	c := require.New(t)
	duration := TimeoutDuration()
	c.Equal(15*time.Second, duration)
}

func TestGet(t *testing.T) {
	c := require.New(t)
	params := url.Values{}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}
	response, err := Get(context.Background(), "http://httpbin.org/get", headers, params)
	c.NoError(err)
	c.NotNil(response)
	c.Equal("200 OK", response.Status)
	c.NoError(response.Body.Close())
}

func TestGetThreadSafe(t *testing.T) {
	c := require.New(t)

	params := url.Values{}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}

	for i := 0; i < 100; i++ {
		go func() {
			response, err := Get(context.Background(), "http://httpbin.org/get", headers, params)
			c.NoError(err)
			c.NotNil(response)
			c.Equal("200 OK", response.Status)
			c.NoError(response.Body.Close())
		}()
	}
}

func TestGetWithParams(t *testing.T) {
	c := require.New(t)
	params := url.Values{"a": {"b"}}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}

	response, err := Get(context.Background(), "http://httpbin.org/get", headers, params)
	c.NoError(err)
	c.NotNil(response)
	c.Equal("200 OK", response.Status)
	c.NoError(response.Body.Close())
}

func TestPost(t *testing.T) {
	c := require.New(t)
	params := url.Values{}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}
	response, err := Post(context.Background(), "http://httpbin.org/post", headers, params)
	c.Nil(err)
	c.NotNil(response)
	c.Equal("200 OK", response.Status)
	c.NoError(response.Body.Close())
}

func TestPostWithParams(t *testing.T) {
	c := require.New(t)
	params := url.Values{"a": {"b"}}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}
	response, err := Post(context.Background(), "http://httpbin.org/post", headers, params)
	c.Nil(err)
	c.NotNil(response)
	c.Equal("200 OK", response.Status)
	c.NoError(response.Body.Close())
}

func TestMock(t *testing.T) {
	c := require.New(t)

	client := InitMock()
	defer client.Close()

	mockedBody := []byte("mocked response")
	mockedResponse := NewMockResponse(http.StatusOK, mockedBody)

	client.AddResponse(http.MethodGet, "http://testingurl.com", mockedResponse)

	req, err := http.NewRequest(http.MethodGet, "http://testingurl.com", nil)
	c.Nil(err)

	res, err := client.Do(req)
	c.Nil(err)

	c.Equal(mockedResponse.StatusCode, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	c.Nil(err)
	c.Equal(mockedBody, body)
	c.NoError(res.Body.Close())
}

func TestMockAddResponseFromFile(t *testing.T) {
	c := require.New(t)

	client := InitMock()
	defer client.Close()

	mockedResponse, err := client.AddResponseFromFile(
		http.MethodGet,
		"http://testingurl.com",
		http.StatusOK,
		"sample.txt",
	)
	c.Nil(err)

	req, err := http.NewRequest(http.MethodGet, "http://testingurl.com", nil)
	c.Nil(err)

	res, err := client.Do(req)
	c.Nil(err)

	c.Equal(mockedResponse.StatusCode, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	c.Nil(err)

	mockedBody, err := os.ReadFile("sample.txt")
	c.Nil(err)
	c.Equal(mockedBody, body)

	_, err = client.AddResponseFromFile(
		http.MethodGet,
		"http://testingurl.com",
		http.StatusOK,
		"not_found.txt",
	)
	c.Equal(ErrMockFailed, err)
	c.NoError(res.Body.Close())
}

func TestAddMockedResponseWithHeaders(t *testing.T) {
	c := require.New(t)

	client, err := New()
	c.Nil(err)

	client.ActivateMock()
	defer client.DeactivateMock()

	client.AddMockedResponseWithHeaders(
		http.MethodGet,
		"http://testingurl.com",
		http.StatusOK,
		"sample.txt",
		http.Header{"Content-Type": {"text/plain"}},
	)
	c.Nil(err)

	r, err := client.Get(
		context.Background(),
		"http://testingurl.com",
		http.Header{},
		url.Values{},
	)
	c.Nil(err)
	c.Equal(r.Header["Content-Type"][0], "text/plain")
	c.NoError(r.Body.Close())
}

type mockReadCloser struct {
	mock.Mock
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockReadCloser) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestMockHTTPError(t *testing.T) {
	c := require.New(t)

	client := InitMock()
	defer client.Close()

	mockedResponse := NewMockResponse(http.StatusBadRequest, nil)

	client.AddResponse("GET", "http://testingurl.com", mockedResponse)

	req, err := http.NewRequest("GET", "http://testingurl.com", nil)
	c.Nil(err)

	res, err := client.Do(req)
	c.Equal(mockedResponse.StatusCode, res.StatusCode)
	c.Equal(err, ErrMockHTTPResponseFailed)

	c.Panics(func() {
		mockRC := mockReadCloser{}
		mockRC.On("Read", mock.AnythingOfType("[]uint8")).Return(0, bytes.ErrTooLarge)

		badResponse := &http.Response{
			Body: &mockRC,
		}
		client.AddResponse("GET", "http://testingurl.com", badResponse)
	})
}

func TestMockError(t *testing.T) {
	c := require.New(t)

	client := InitMock()
	defer client.Close()

	req, err := http.NewRequest("GET", "http://testingurl.com", nil)
	c.Nil(err)

	res, err := client.Do(req)
	c.Nil(res)
	c.Equal(err, ErrMockNotFound)

	if err == nil {
		c.NoError(res.Body.Close())
	}
}

func TestNewWithDoer(t *testing.T) {
	c := require.New(t)

	client := NewWithDoer(&http.Client{})
	c.NotNil(client)
	c.NotEmpty(client.DefaultGetHeaders["User-Agent"])
}

func TestNewFailing(t *testing.T) {
	c := require.New(t)

	prevCookieJarFunc := newCookieJarFunc
	newCookieJarFunc = func(o *cookiejar.Options) (*cookiejar.Jar, error) {
		return nil, errors.New("failed")
	}

	defer func() {
		newCookieJarFunc = prevCookieJarFunc
	}()

	client, err := New()
	c.Error(err)
	c.Nil(client)
}

func TestNewWithXRay(t *testing.T) {
	c := require.New(t)

	client, err := NewWithXRay()
	c.NoError(err)
	c.NotNil(client)
	c.NotEmpty(client.DefaultGetHeaders["User-Agent"])

	prevCookieJarFunc := newCookieJarFunc
	newCookieJarFunc = func(o *cookiejar.Options) (*cookiejar.Jar, error) {
		return nil, errors.New("failed")
	}

	defer func() {
		newCookieJarFunc = prevCookieJarFunc
	}()

	_, err = NewWithXRay()
	c.Error(err)
}

func TestNewWithLogProperties(t *testing.T) {
	c := require.New(t)

	propertiesToLog := logger.MapObject("some_object", map[string]interface{}{"s_some_property": "hello"})

	client, err := NewWithLogProperties(propertiesToLog)
	c.Nil(err)
	c.Equal([]logger.Object{propertiesToLog}, client.propertiesToLog)

	buffer := bytes.NewBuffer(nil)
	log := logger.New("test")
	log.Output = buffer

	ctx := logger.Set(context.Background(), log)

	response, err := client.Get(ctx, "http://httpbin.org/get", nil, nil)
	c.Nil(err)

	defer func() {
		err := response.Body.Close()
		c.Nil(err)
	}()

	c.Contains(buffer.String(), "\"some_object\":{\"s_some_property\":\"hello\"}}")
}

func TestNewWithOptions(t *testing.T) {
	c := require.New(t)

	propertiesToLog := logger.MapObject("some_object", map[string]interface{}{"s_some_property": "hello"})

	proxy, err := url.Parse("http://u:pass@testingurl.com")
	c.Nil(err)

	clientOptions := &NewClientOptions{
		PropertiesToLog: []logger.Object{propertiesToLog},
		UseXRay:         true,
		StaticProxy:     proxy,
	}

	client, err := NewWithOptions(clientOptions)
	c.Nil(err)
	c.Equal(proxy, client.staticProxy)
	c.Equal([]logger.Object{propertiesToLog}, client.propertiesToLog)

	buffer := bytes.NewBuffer(nil)
	log := logger.New("test")
	log.Output = buffer

	ctx, segment := xray.BeginSegment(context.Background(), "httpbin.org")
	defer segment.Close(nil)

	ctx = logger.Set(ctx, log)

	response, err := client.Get(ctx, "http://httpbin.org/get", nil, nil)
	c.Nil(err)

	defer func() {
		err := response.Body.Close()
		c.Nil(err)
	}()

	c.Contains(buffer.String(), "\"some_object\":{\"s_some_property\":\"hello\"}}")
	c.Contains(buffer.String(), "\"s_proxy_used\":\"http://u:d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1@testingurl.com\"}")
}

func TestNewWithOptionsAndForceHttpsRedirect(t *testing.T) {
	c := require.New(t)

	forceHTTPSRedirect = true

	defer func() {
		forceHTTPSRedirect = false
	}()

	clientOptions := &NewClientOptions{}

	client, err := NewWithOptions(clientOptions)
	c.Nil(err)
	c.NotNil(client.HTTPClient.CheckRedirect)

	ctx := context.Background()

	response, err := client.Get(ctx, "http://google.com", nil, nil)
	c.Nil(err)

	defer func() {
		err = response.Body.Close()
		c.Nil(err)
	}()

	forceHTTPSRedirect = false
	clientOptions = &NewClientOptions{}

	client, err = NewWithOptions(clientOptions)
	c.Nil(err)
	c.Nil(client.HTTPClient.CheckRedirect)

	response, err = client.Get(ctx, "http://google.com", nil, nil)
	c.Nil(err)

	defer func() {
		err := response.Body.Close()
		c.Nil(err)
	}()
}

func TestSendRequest(t *testing.T) {
	c := require.New(t)

	response, err := SendRequest(context.Background(), "GET", "https://truora.com", http.Header{}, "")
	c.Nil(err)
	c.Equal(200, response.StatusCode)
	c.NoError(response.Body.Close())
}

func TestSendRawRequest(t *testing.T) {
	c := require.New(t)

	request, err := http.NewRequest(http.MethodGet, "https://truora.com", bytes.NewBuffer([]byte("")))
	c.Nil(err)

	response, err := Default.SendRawRequest(context.Background(), request)
	c.Nil(err)
	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(response.Body.Close())
}

func Test_retrier(t *testing.T) {
	c := require.New(t)

	c.Equal(0*time.Millisecond, retrier(-1))
	c.Equal(40*time.Millisecond, retrier(1))
	c.Equal(80*time.Millisecond, retrier(2))
	c.Equal(120*time.Millisecond, retrier(3))
}

func TestActivateDeactivateMock(t *testing.T) {
	c := require.New(t)

	ActivateMock()

	res, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.True(strings.Contains(err.Error(), "Get \"https://api.truora.com\": no responder found"))

	if err == nil {
		c.NoError(res.Body.Close())
	}

	DeactivateMock()
}

func TestAddMockedResponseFromFile(t *testing.T) {
	c := require.New(t)

	ActivateMock()

	defer DeactivateMock()

	AddMockedResponseFromFile("GET", "https://api.truora.com", http.StatusCreated, "sample.txt")

	response, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Nil(err)
	c.NotNil(response)
	c.Equal(http.StatusCreated, response.StatusCode)
	c.NoError(response.Body.Close())

	c.Panics(func() {
		AddMockedResponseFromFile("GET", "https://api.truora.com", http.StatusCreated, "not_found.txt")
	})
}

func TestAddMultipleMockedResponsesSuccess(t *testing.T) {
	c := require.New(t)

	ActivateMock()

	defer DeactivateMock()

	AddMultipleMockedResponses(http.MethodGet, "https://api.truora.com", http.StatusOK, []string{
		"sample.txt",
		"sample.txt",
	})

	response, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Nil(err)
	c.NotNil(response)
	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(response.Body.Close())

	response, err = Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Nil(err)
	c.NotNil(response)
	c.Equal(http.StatusOK, response.StatusCode)
	c.NoError(response.Body.Close())
}

func TestAddMultipleMockedResponsesFailed(t *testing.T) {
	c := require.New(t)

	ActivateMock()

	defer DeactivateMock()

	AddMultipleMockedResponses(http.MethodGet, "https://api.truora.com", http.StatusOK, []string{
		"sample.txt",
		"sample.txt",
	})

	response1, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Nil(err)
	c.NotNil(response1)
	c.Equal(http.StatusOK, response1.StatusCode)
	c.NoError(response1.Body.Close())

	response2, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Nil(err)
	c.NotNil(response2)
	c.Equal(http.StatusOK, response2.StatusCode)
	c.NoError(response2.Body.Close())

	response3, err := Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	if response3 != nil {
		c.NoError(response3.Body.Close())
		c.Nil(response3)
	}

	c.Nil(response3)
	c.Error(ErrResponseNotFound, err)
}

func TestSetCookies(t *testing.T) {
	c := require.New(t)

	client, err := New()
	c.Nil(err)

	err = client.SetCookies("https://api.truora.com", "c1=1;c2=2;")
	c.Nil(err)

	cookies := client.GetCookies("https://api.truora.com")
	c.NotEmpty(cookies)
	c.Equal(cookies, "c1=1;c2=2;")
}

func TestClearCookies(t *testing.T) {
	c := require.New(t)

	client, err := New()
	c.Nil(err)

	err = client.SetCookies("https://api.truora.com", "c1=1;c2=2;")
	c.Nil(err)

	cookies := client.GetCookies("https://api.truora.com")
	c.NotEmpty(cookies)
	c.Equal(cookies, "c1=1;c2=2;")

	err = client.ClearCookies("https://api.truora.com")
	c.Nil(err)

	cookies = client.GetCookies("https://api.truora.com")
	c.Empty(cookies)

	// test thread safety
	var wg sync.WaitGroup

	workers := 2
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			err := client.SetCookies("https://api.truora.com", "c1=1;c2=2;")
			c.Nil(err)

			err = client.ClearCookies("https://api.truora.com")
			c.Nil(err)
		}()
	}

	wg.Wait()
}

func TestSetCookiesError(t *testing.T) {
	c := require.New(t)

	client, err := New()
	c.Nil(err)

	cookies := client.GetCookies("	")
	c.Empty(cookies)
}

func TestExcecAndLog_FailedInResponseEncodeDeduce(t *testing.T) {
	c := require.New(t)

	cli, err := New()
	c.Nil(err)

	cli.ActivateMock()
	defer cli.DeactivateMock()

	httpmock.RegisterResponder("GET", "http://testingurl.com", func(*http.Request) (*http.Response, error) {
		return &http.Response{
			Header: http.Header{
				"Content-Type": []string{
					"text/html; charset=unknown;",
				},
			},
			Body: io.NopCloser(strings.NewReader("content")),
		}, nil
	})

	u, err := url.Parse("http://testingurl.com")
	c.Nil(err)

	request := &http.Request{
		URL: u,
	}

	log := logger.New("log")
	loggerWriter := bytes.NewBufferString("")
	log.Output = loggerWriter

	ctx := context.Background()
	ctx = logger.Set(ctx, log)

	res, err := cli.execAndLog(ctx, request, nil)
	c.Nil(err)
	c.Contains(loggerWriter.String(), "deduce_response_encoding_failed")
	c.NoError(res.Body.Close())
}

func TestExcecAndLog_CharsetEncodingDistinctThanUTF8(t *testing.T) {
	c := require.New(t)

	cli, err := New()
	c.Nil(err)

	cli.ActivateMock()
	defer cli.DeactivateMock()

	httpmock.RegisterResponder("GET", "http://testingurl.com", func(*http.Request) (*http.Response, error) {
		return &http.Response{
			Header: http.Header{
				"Content-Type": []string{
					"text/html; charset=iso-8859-1;",
				},
			},
			Body: io.NopCloser(strings.NewReader("content")),
		}, nil
	})

	u, err := url.Parse("http://testingurl.com")
	c.Nil(err)

	request := &http.Request{
		URL: u,
	}

	log := logger.New("log")
	loggerWriter := bytes.NewBufferString("")
	log.Output = loggerWriter

	ctx := context.Background()
	ctx = logger.Set(ctx, log)

	res, err := cli.execAndLog(ctx, request, nil)
	c.Nil(err)
	c.Contains(loggerWriter.String(), "unknown_response_encoding", "iso-8859-1" /* should also log the unknown encoding charset*/)
	c.NoError(res.Body.Close())
}

func TestAddMockedResponseWithError(t *testing.T) {
	c := require.New(t)

	cli, err := New()
	c.Nil(err)

	cli.ActivateMock()
	defer cli.DeactivateMock()

	expected := cli.AddMockedResponseWithError(http.MethodGet, "https://api.truora.com", errors.New("no responder found"))

	res, err := cli.Get(context.Background(), "https://api.truora.com", http.Header{}, url.Values{})
	c.Equal(expected, err.Error())

	if err == nil {
		c.NoError(res.Body.Close())
	}
}

func TestGetWithProxy(t *testing.T) {
	c := require.New(t)

	ctx := context.Background()

	proxy, err := url.Parse("http://testingurl.com")
	c.NoError(err)

	client, err := NewWithProxy(proxy)
	c.NoError(err)

	c.Equal(client.staticProxy, proxy)

	params := url.Values{}
	headers := http.Header{
		"Host": {"httpbin.org"},
	}

	buffer := bytes.NewBuffer(nil)
	log := logger.New("test")
	log.Output = buffer

	ctx = logger.Set(ctx, log)

	response, err := client.Get(ctx, "http://httpbin.org/get", headers, params)
	c.Nil(err)

	defer func() {
		err := response.Body.Close()
		c.Nil(err)
	}()

	c.Contains(buffer.String(), "\"s_proxy_used\":\"http://testingurl.com\"}")
}

func TestLogRequestBlocked(t *testing.T) {
	c := require.New(t)

	Default.ActivateMock()
	defer Default.DeactivateMock()

	Default.AddMockedResponse("GET", "test.page", http.StatusOK, "response")

	ctx := context.Background()
	log := logger.New("test")
	buff := bytes.NewBufferString("")
	log.Output = buff
	ctx = logger.Set(ctx, log)

	err := Default.LogRequestBlocked(ctx, nil)
	c.Equal(ErrInvalidResponse, err)

	proxy, err := url.Parse("proxy.url")
	c.Nil(err)

	Default.staticProxy = proxy

	response, err := Default.Get(ctx, "test.page", nil, nil)
	c.Nil(err)
	c.NoError(response.Body.Close())

	err = Default.LogRequestBlocked(ctx, response)
	c.Nil(err)

	c.Contains(buff.String(), "proxy.url")
	c.Contains(buff.String(), "http_request_blocked")
}

func TestForbiddenResponse(t *testing.T) {
	c := require.New(t)

	Default.ActivateMock()
	defer Default.DeactivateMock()

	Default.AddMockedResponse("GET", "test.page", http.StatusForbidden, "response")

	ctx := context.Background()
	log := logger.New("test")
	buff := bytes.NewBufferString("")
	log.Output = buff
	ctx = logger.Set(ctx, log)

	response, err := Default.Get(ctx, "test.page", nil, nil)
	c.Nil(err)
	c.Equal(403, response.StatusCode)

	c.NoError(response.Body.Close())

	c.Contains(buff.String(), "http_request_blocked")
}

func TestForbiddenResponseCloudFlare(t *testing.T) {
	c := require.New(t)

	client, err := New()
	c.Nil(err)

	client.ActivateMock()
	defer client.DeactivateMock()

	ctx := context.Background()
	log := logger.New("test")
	buff := bytes.NewBufferString("")
	log.Output = buff
	ctx = logger.Set(ctx, log)

	client.AddMockedResponseWithHeaders(
		http.MethodGet,
		"http://testingurl.com",
		http.StatusForbidden,
		"sample.txt",
		http.Header{"Server": {"cloudflare"}},
	)
	c.Nil(err)

	response, err := client.Get(
		ctx,
		"http://testingurl.com",
		http.Header{},
		url.Values{},
	)
	c.Nil(err)
	c.Equal(403, response.StatusCode)

	c.NoError(response.Body.Close())

	c.Contains(buff.String(), "cloudflare_challenge_detected")
}
