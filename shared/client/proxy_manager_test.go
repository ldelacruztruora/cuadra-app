package client

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/stretchr/testify/require"
)

func TestNewProxyManager(t *testing.T) {
	c := require.New(t)

	err := os.Setenv("TRUORA_HTTP_PROXY", "https://myproxy.com;http://testingproxy.com")
	c.NoError(err)
	err = os.Setenv("USE_PROXY_ROTATION", "true")
	c.NoError(err)

	defer func() {
		_ = os.Unsetenv("TRUORA_HTTP_PROXY")
		_ = os.Unsetenv("USE_PROXY_ROTATION")
	}()

	proxyManager := NewProxyManager()
	c.True(proxyManager.useProxyRotation)
	c.Equal(2, len(proxyManager.proxyURLs))
}

func TestNewProxyManagerUseProxyRotationFalse(t *testing.T) {
	c := require.New(t)

	proxyManager := NewProxyManager()
	c.False(proxyManager.useProxyRotation)
}

func TestProxyForRequestWithVariable(t *testing.T) {
	c := require.New(t)

	proxy, _ := url.Parse("https://myproxy.com")
	manager := &ProxyManager{
		proxyURLs: []*url.URL{proxy},
	}

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)
	c.NoError(req.Body.Close())

	request, _ := manager.SetProxyInRequest(req, proxy)
	c.NoError(req.Body.Close())

	proxyURL, err := manager.ProxyForRequest(request)
	c.NoError(err)

	c.Equal("myproxy.com", proxyURL.Host)
}

func TestProxyForRequestWithVariableForAmazon(t *testing.T) {
	c := require.New(t)

	proxy, _ := url.Parse("https://myproxy.com")
	manager := &ProxyManager{
		proxyURLs: []*url.URL{proxy},
	}

	req, err := http.NewRequest("GET", "http://sqs.us-east-1.amazonaws.com", strings.NewReader(""))
	c.Nil(err)
	proxyURL, err := manager.ProxyForRequest(req)
	c.NoError(err)
	c.Nil(proxyURL)
	c.NoError(req.Body.Close())
}

func TestProxyForRequestWithNoVariable(t *testing.T) {
	c := require.New(t)

	manager := &ProxyManager{
		proxyURLs: []*url.URL{},
	}

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.Nil(err)
	proxyURL, err := manager.ProxyForRequest(req)
	c.NoError(err)
	c.Nil(proxyURL)
	c.NoError(req.Body.Close())
}

func TestProxyForRequestWithLambdaContext(t *testing.T) {
	c := require.New(t)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "2a034f70-79b0-4de7-a082-af37189bf3ba",
	})

	proxy, _ := url.Parse("https://u:p@myproxy.com")
	manager := &ProxyManager{
		proxyURLs: []*url.URL{proxy, proxy, proxy},
	}

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	req, _ = manager.SetProxyInRequest(req.WithContext(ctx), nil)

	proxyURL, err := manager.ProxyForRequest(req)
	c.NoError(err)
	c.NoError(req.Body.Close())

	c.Equal("myproxy.com", proxyURL.Host)
}

func TestSetProxyInRequestWithLambdaContext(t *testing.T) {
	c := require.New(t)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "2a034f70-79b0-4de7-a082-af37189bf3ba",
	})

	expectedProxyURL, _ := url.Parse("https://myproxy.com")

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	req = req.WithContext(ctx)

	manager := &ProxyManager{}
	req, _ = manager.SetProxyInRequest(req, expectedProxyURL)

	proxyURL, err := manager.ProxyForRequest(req)
	c.NoError(err)
	c.NoError(req.Body.Close())
	c.Equal(expectedProxyURL, proxyURL)
}

func TestSetProxyInRequestNilWithProxyRotation(t *testing.T) {
	c := require.New(t)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "2a034f70-79b0-4de7-a082-af37189bf3ba",
	})

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	req = req.WithContext(ctx)

	manager := &ProxyManager{useProxyRotation: true}
	req, proxyURL := manager.SetProxyInRequest(req, nil)

	c.Nil(proxyURL)
	c.NoError(req.Body.Close())
}

func TestGetProxyFromRequestContextNil(t *testing.T) {
	c := require.New(t)

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	proxy, err := DefaultProxyManager.ProxyForRequest(req)
	c.Nil(err)
	c.Nil(proxy)
	c.NoError(req.Body.Close())
}

func TestGetProxyFromRequestContext(t *testing.T) {
	c := require.New(t)

	expectedProxyURL, err := url.Parse("https://myproxy.com")
	c.NoError(err)

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	ctx := context.Background()
	req = req.WithContext(ctx)

	req, _ = DefaultProxyManager.SetProxyInRequest(req, expectedProxyURL)

	proxyURL, err := DefaultProxyManager.ProxyForRequest(req)
	c.Nil(err)
	c.Equal(expectedProxyURL, proxyURL)
	c.NoError(req.Body.Close())
}

func TestLoadProxiesFromEnv(t *testing.T) {
	c := require.New(t)

	err := os.Setenv("TRUORA_HTTP_PROXY", "https://myproxy.com;http://testingproxy.com")
	c.Nil(err)

	defer func() {
		_ = os.Unsetenv("TRUORA_HTTP_PROXY")
	}()

	proxyManager := NewProxyManager()
	c.Len(proxyManager.proxyURLs, 2)

	err = os.Setenv("TRUORA_HTTP_PROXY", "https://new.com")
	c.Nil(err)

	proxyManager.SetProxiesFromEnv()
	c.Len(proxyManager.proxyURLs, 1)
	c.Equal("https://new.com", proxyManager.proxyURLs[0].String())

	err = os.Setenv("TRUORA_HTTP_PROXY", "")
	c.Nil(err)

	proxyManager.SetProxiesFromEnv()
	c.Len(proxyManager.proxyURLs, 0)
}

func BenchmarkProxyForRequestWithContext(b *testing.B) {
	c := require.New(b)
	ctx := lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
		AwsRequestID: "2a034f70-79b0-4de7-a082-af37189bf3ba",
	})

	proxy, _ := url.Parse("https://myproxy.com")
	manager := &ProxyManager{
		proxyURLs: []*url.URL{proxy, proxy, proxy},
	}

	req, err := http.NewRequest("GET", "http://testingurl.com", strings.NewReader(""))
	c.NoError(err)

	for i := 0; i < b.N; i++ {
		_, err := manager.ProxyForRequest(req.WithContext(ctx))
		c.NoError(err)
	}

	c.NoError(req.Body.Close())
}
