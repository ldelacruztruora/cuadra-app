// Package client decleare client struct and functions to manage and use it in the http requests also manage proxy functions
package client

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/jarcoal/httpmock"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"bitbucket.org/truora/scrap-services/logger"
	"bitbucket.org/truora/scrap-services/shared/env"
)

// Client defines an http client
type Client struct {
	client heimdall.Client

	// DefaultPostHeaders of a POST request
	DefaultPostHeaders map[string]string

	// DefaultGetHeaders of a GET request
	DefaultGetHeaders map[string]string

	HTTPClient *http.Client // NOTE: this can be nil

	// staticProxy sets a static proxy.
	// It is chosen over a random one
	staticProxy *url.URL

	propertiesToLog []logger.Object
}

// NewClientOptions holds options for client creation
type NewClientOptions struct {
	UseXRay         bool
	PropertiesToLog []logger.Object
	StaticProxy     *url.URL
	CheckRedirect   func(req *http.Request, via []*http.Request) error
}

var (
	// ErrNoHTTPClient when http client is not present
	ErrNoHTTPClient = errors.New("http client is not present")
	// ErrResponseNotFound when try get a response not avaible
	ErrResponseNotFound = errors.New("response not found")
	// ErrInvalidResponse when the response passed is invalid
	ErrInvalidResponse = errors.New("invalid http response")
	newCookieJarFunc   = cookiejar.New
	rotationActive     = env.GetBool("USE_PROXY_ROTATION", false)
)

// NewWithXRay creates a new http client instrumented with XRay with pool and retries.
func NewWithXRay() (*Client, error) {
	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	return NewWithDoer(xray.Client(client)), nil
}

// New creates a new http client with pool and retrier
func New() (*Client, error) {
	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	return NewWithDoer(client), nil
}

// NewWithLogProperties creates a new http client and logs the specified log properties with each request
func NewWithLogProperties(propertiestoLog ...logger.Object) (*Client, error) {
	httpClient, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	client := NewWithDoer(httpClient)
	client.propertiesToLog = propertiestoLog

	return client, nil
}

// NewWithProxy creates a new http client with proxy
func NewWithProxy(proxy *url.URL) (*Client, error) {
	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	cli := NewWithDoer(client)
	cli.staticProxy = proxy

	return cli, nil
}

// NewWithOptions creates a new http client with proxy
func NewWithOptions(options *NewClientOptions) (*Client, error) {
	client, err := newHTTPClient()
	if err != nil {
		return nil, err
	}

	cli := NewWithDoer(client)

	if options == nil {
		return cli, nil
	}

	if options.UseXRay {
		cli = NewWithDoer(xray.Client(client))
	}

	if options.StaticProxy != nil {
		cli.staticProxy = options.StaticProxy
	}

	if options.PropertiesToLog != nil && len(options.PropertiesToLog) > 0 {
		cli.propertiesToLog = options.PropertiesToLog
	}

	if forceHTTPSRedirect {
		cli.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			req.URL.Scheme = "https"

			return nil
		}
	}

	return cli, nil
}

func newHTTPClient() (*http.Client, error) {
	cookieJar, err := newCookieJarFunc(nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:       cookieJar,
		Transport: DefaultRoundTripper,
		Timeout:   httpClientTimeout,
	}

	return client, nil
}

func retrier(retry int) time.Duration {
	if retry <= 0 {
		return 0 * time.Millisecond
	}

	return time.Duration(10*(retry<<2)) * time.Millisecond
}

// NewWithDoer creates a new client with the given a heimdall.Doer
func NewWithDoer(doer heimdall.Doer) *Client {
	heimdallClient := httpclient.NewClient(
		httpclient.WithHTTPTimeout(httpClientTimeout),
		httpclient.WithHTTPClient(doer),
		httpclient.WithRetrier(heimdall.NewRetrierFunc(retrier)),
		httpclient.WithRetryCount(httpClientRetries),
	)
	userAgent := userAgents[rand.Intn(len(userAgents))] // #nosec G404

	client, _ := doer.(*http.Client)

	return &Client{
		client: heimdallClient,
		DefaultPostHeaders: map[string]string{
			"User-Agent":   userAgent,
			"Content-Type": "application/x-www-form-urlencoded",
			"Accept":       "*/*",
		},
		DefaultGetHeaders: map[string]string{
			"User-Agent": userAgent,
		},
		HTTPClient: client,
	}
}

// Get does an http GET request
func (c *Client) Get(ctx context.Context, url string, headers http.Header, params url.Values) (*http.Response, error) {
	paramsStr := ""
	urlWithParams := url

	if len(params) != 0 {
		paramsStr = params.Encode()
		urlWithParams = fmt.Sprintf("%s?%s", url, paramsStr)
	}

	return c.SendRequest(ctx, "GET", urlWithParams, headers, "")
}

// Post does an http POST request
func (c *Client) Post(ctx context.Context, url string, headers http.Header, params url.Values) (*http.Response, error) {
	paramsStr := ""
	if len(params) != 0 {
		paramsStr = params.Encode()
	}

	return c.SendRequest(ctx, "POST", url, headers, paramsStr)
}

// SendRequest sends a http request
func (c *Client) SendRequest(ctx context.Context, method string, url string, headers http.Header, body string) (*http.Response, error) {
	request, err := http.NewRequest(method, url, buildBodyReader(body))
	if err != nil {
		return nil, err
	}

	request = request.WithContext(ctx)

	header := c.defaultHeader(method)
	for key, value := range headers {
		header[key] = value
	}

	request.Header = header

	request, proxy := DefaultProxyManager.SetProxyInRequest(request, c.staticProxy)

	return c.execAndLog(ctx, request, proxy)
}

// SendRawRequest sends a http request as defined in a http.Request struct
func (c *Client) SendRawRequest(ctx context.Context, request *http.Request) (*http.Response, error) {
	request = request.WithContext(ctx)

	originalHeaders := request.Header
	defaultHeaders := c.defaultHeader(request.Method)

	for key := range defaultHeaders {
		if value, ok := defaultHeaders[key]; !ok {
			originalHeaders.Set(key, value[0])
		}
	}

	request.Header = originalHeaders

	request, proxy := DefaultProxyManager.SetProxyInRequest(request, c.staticProxy)

	return c.execAndLog(ctx, request, proxy)
}

func buildBodyReader(body string) io.Reader {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	return bodyReader
}

func (c *Client) defaultHeader(method string) http.Header {
	header := http.Header{}

	if method == "POST" || method == "PUT" {
		for key, value := range c.DefaultPostHeaders {
			header.Set(key, value)
		}
	} else {
		for key, value := range c.DefaultGetHeaders {
			header.Set(key, value)
		}
	}

	return header
}

func (c *Client) execAndLog(ctx context.Context, request *http.Request, proxy *url.URL) (*http.Response, error) {
	t := time.Now()

	proxyURL := ""
	if proxy != nil {
		// we only need the host in the logs
		proxyURL = hashBasicAuth(proxy)
	}

	response, err := c.client.Do(request)

	if proxy == nil && rotationActive {
		proxyURL = NoneProxy
	}

	if response != nil {
		logger.Get(ctx).Info(ctx, "http_request_performed", logger.ThreeMonths, append(c.propertiesToLog, []logger.Object{
			logger.HTTPRequest(request),
			logger.HTTPResponse(response, time.Since(t)),
			logger.MapObject("proxy_info", map[string]interface{}{
				"s_proxy_used": proxyURL,
			}),
		}...))
	}

	if err != nil {
		logger.Get(ctx).Warning(ctx, "http_request_failed", logger.ThreeMonths, append(c.propertiesToLog, []logger.Object{
			logger.HTTPRequest(request),
			logger.ErrObject(err),
			logger.MapObject("proxy_info", map[string]interface{}{
				"s_proxy_used": proxyURL,
			}),
		}...))

		return nil, err
	}

	// log response charset other than default
	checkAndLogResponseCharset(ctx, response, request)

	return response, logBasedOnResponseStatus(ctx, response, c)
}

func logBasedOnResponseStatus(ctx context.Context, response *http.Response, client *Client) error {
	if response.StatusCode == http.StatusForbidden && response.Header.Get("Server") == "cloudflare" {
		logger.Get(ctx).Warning(ctx, "cloudflare_challenge_detected", logger.ThreeMonths, append(client.propertiesToLog,
			logger.HTTPRequest(response.Request)))
	}

	// valid status code is not forbidden
	if response.StatusCode == http.StatusForbidden {
		return client.LogRequestBlocked(ctx, response)
	}

	return nil
}

func checkAndLogResponseCharset(ctx context.Context, response *http.Response, initialRequest *http.Request) {
	if initialRequest == nil {
		return
	}

	// can not deduce reponse charset, log it and return
	name, err := deduceCharset(response)
	if err != nil {
		logger.Get(ctx).Warning(ctx, "deduce_response_encoding_failed", logger.OneDay, []logger.Object{
			logger.HTTPRequest(initialRequest),
			logger.ErrObject(err),
		})

		return
	}

	// valid charset, nothing to log
	if name == defaultCharset {
		return
	}

	// if encoding and charset was correctly deduced, and charset is not default, log it
	logger.Get(ctx).Warning(ctx, "unknown_response_encoding", logger.OneDay, []logger.Object{
		logger.HTTPRequest(initialRequest),
		logger.ErrObject(fmt.Errorf("unknown response encoding '%s' in page %s", name, initialRequest.URL.String())),
	})
}

// ActivateMock activates the mock
func (c *Client) ActivateMock() {
	if c.HTTPClient != nil {
		httpmock.ActivateNonDefault(c.HTTPClient)
	}
}

// DeactivateMock deactivates the mock
func (c *Client) DeactivateMock() {
	if c.HTTPClient != nil {
		httpmock.DeactivateAndReset()
	}
}

// AddMockedResponseFromFile adds a mocked response given a file path relative to the test file
func (c *Client) AddMockedResponseFromFile(method string, url string, statusCode int, filePath string) {
	data, err := os.ReadFile(filePath) // #nosec
	if err != nil {
		panic(err)
	}

	c.AddMockedResponse(method, url, statusCode, string(data))
}

// AddMockedResponse adds a mocked response given its content
func (c *Client) AddMockedResponse(method string, url string, statusCode int, content string) {
	responder := httpmock.NewStringResponder(statusCode, content)
	httpmock.RegisterResponder(method, url, responder)
}

// AddMultipleMockedResponses add a mocked response given one to one from each file
func (c *Client) AddMultipleMockedResponses(method string, url string, statusCode int, responseList []string) {
	var mutex = sync.Mutex{}

	nextResponseIndex := 0
	responseFunction := func(req *http.Request) (*http.Response, error) {
		mutex.Lock()
		defer mutex.Unlock()

		if nextResponseIndex >= len(responseList) {
			return nil, ErrResponseNotFound
		}

		data, err := os.ReadFile(responseList[nextResponseIndex]) // #nosec
		if err != nil {
			panic(err)
		}

		req.Response = httpmock.NewStringResponse(statusCode, string(data))

		nextResponseIndex++

		return req.Response, nil
	}

	httpmock.RegisterResponder(method, url, responseFunction)
}

// AddMockedResponseWithHeaders adds a mocked response given its content and headers
func (c *Client) AddMockedResponseWithHeaders(method string, url string, statusCode int, content string, headers http.Header) {
	response := &http.Response{
		Status:     strconv.Itoa(statusCode),
		StatusCode: statusCode,
		Body:       httpmock.NewRespBodyFromBytes([]byte(content)),
		Header:     headers,
	}

	responder := httpmock.ResponderFromResponse(response)
	httpmock.RegisterResponder(method, url, responder)
}

// AddMockedResponseWithError adds a mocked response with an error
func (c *Client) AddMockedResponseWithError(method string, url string, err error) string {
	responder := httpmock.NewErrorResponder(err)
	httpmock.RegisterResponder(method, url, responder)
	// it generates the message error with heimdall's format
	// which contains the error message of every retry done by heimdall

	caser := cases.Title(language.Und)

	errs := []string{}
	titleizeMethod := caser.String(strings.ToLower(method))

	for i := 0; i <= httpClientRetries; i++ {
		errs = append(errs, titleizeMethod+" \""+url+"\": "+err.Error())
	}

	return strings.Join(errs, ", ")
}

// SetCookies fill the cookie jar with the given rawCookies
func (c *Client) SetCookies(rawURL string, rawCookies string) error {
	if c.HTTPClient == nil {
		return nil
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	header := http.Header{}
	header.Add("Cookie", rawCookies)

	respond := http.Request{
		Header: header,
	}

	cookies := respond.Cookies()
	c.HTTPClient.Jar.SetCookies(u, cookies)

	return nil
}

// GetCookies obtain cookies from the cookie jar
func (c *Client) GetCookies(rawURL string) string {
	if c.HTTPClient == nil {
		return ""
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		log := logger.NewLambdaLogger()
		log.Error(context.Background(), "url_parsel_failed-client", logger.ThreeMonths, []logger.Object{
			logger.ErrObject(err),
		})

		return ""
	}

	cookies := c.HTTPClient.Jar.Cookies(u)
	out := ""

	for _, cookie := range cookies {
		out += cookieToString(cookie) + ";"
	}

	return out
}

// ClearCookies remove all cookies from the cookie jar
func (c *Client) ClearCookies(rawURL string) error {
	if c.HTTPClient == nil {
		return ErrNoHTTPClient
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	cookies := c.HTTPClient.Jar.Cookies(u)
	for _, cookie := range cookies {
		cookie.MaxAge = -1
	}

	c.HTTPClient.Jar.SetCookies(u, cookies)

	return nil
}

func cookieToString(cookie *http.Cookie) string {
	return cookie.Name + "=" + cookie.Value
}

func hashBasicAuth(u *url.URL) string {
	newURL := url.URL{
		Scheme:      u.Scheme,
		Opaque:      u.Opaque,
		Host:        u.Host,
		Path:        u.Path,
		RawPath:     u.RawPath,
		ForceQuery:  u.ForceQuery,
		RawQuery:    u.RawQuery,
		Fragment:    u.Fragment,
		RawFragment: u.RawFragment,
		User:        u.User,
	}

	if u.User != nil {
		pwd, ok := u.User.Password()
		if ok {
			sum := sha256.Sum256([]byte(pwd))
			newURL.User = url.UserPassword(u.User.Username(), fmt.Sprintf("%x", sum))
		}
	}

	return newURL.String()
}

// LogRequestBlocked logs request blocked event with request information obtained from the response
// For use when a request block is manually detected within the request response.
// The event logged is used by the proxy health system for proxy selection.
func (c *Client) LogRequestBlocked(ctx context.Context, response *http.Response) error {
	proxyURL := ""

	if response == nil {
		return ErrInvalidResponse
	}

	proxy, ok := response.Request.Context().Value(proxyCtxKey).(*url.URL)
	if ok {
		proxyURL = hashBasicAuth(proxy)
	}

	if proxy == nil && rotationActive {
		proxyURL = NoneProxy
	}

	logger.Get(ctx).Info(ctx, "http_request_blocked", logger.ThreeMonths, append(c.propertiesToLog, []logger.Object{
		logger.HTTPRequest(response.Request),
		logger.MapObject("proxy_info", map[string]interface{}{
			"s_proxy_used": proxyURL,
		}),
	}...))

	return nil
}
