package httpclient

import (
	"errors"
	"golang.org/x/net/context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"runtime"
	"time"

	"github.com/gojek/heimdall"
	"github.com/gojek/heimdall/httpclient"
	"github.com/gojek/heimdall/hystrix"
	"golang.org/x/net/publicsuffix"
)

type Client struct {
	httpClients map[string]ClientRequestMapping
}

// ClientRequestMapping provides a container for heimdall client and associated RequestConfig.
type ClientRequestMapping struct {
	heimdallClient heimdall.Client
	requestConfig  *RequestConfig
}

// ConfigureHTTPClient receives RequestConfigs and initializes one http client per RequestConfig.
// It creates heimdall http or hystrix client based on the configuration provided in RequestConfig.
// Returns the instance of Client
func ConfigureHTTPClient(requestConfigs ...*RequestConfig) *Client {
	httpClients := make(map[string]ClientRequestMapping)

	for _, requestConfig := range requestConfigs {
		if requestConfig != nil {
			clientRequestMapping :=
				ClientRequestMapping{
					heimdallClient: buildHTTPClient(requestConfig),
					requestConfig:  requestConfig,
				}
			httpClients[requestConfig.name] = clientRequestMapping
		}
	}

	client := Client{
		httpClients: httpClients,
	}

	return &client
}

// Request receives Request param to execute. It will fetch the right http client for given Request name
// and use it to execute based on attributes provided in Request
// It returns http.Response and error
func (c *Client) Request(request *Request) (*http.Response, error) {
	client := c.httpClients[request.name]

	if request.url == "" {
		request.url = client.requestConfig.url
	}

	req, err := getRequest(request.ctx, client.requestConfig.method, request.url, request.queryParams,
		request.headerParams, request.body)

	if err != nil {
		return nil, err
	}

	return client.heimdallClient.Do(req)
}

// This is an internal method to form the http.Request based on various parameters.
func getRequest(ctx context.Context, method string, url string, queryParams map[string]string,
	headerParams map[string]string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		request = request.WithContext(ctx)
	}

	if queryParams != nil {
		q := request.URL.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		request.URL.RawQuery = q.Encode()
	}

	for k, v := range headerParams {
		request.Header.Add(k, v)
	}

	return request, err
}

// Internal method to build http or hystrix client based on settings provided in RequestConfig.
// It will create hystrix client if hystrixConfig is provided else it will provide httpclient.
func buildHTTPClient(requestConfig *RequestConfig) heimdall.Client {
	if requestConfig.hystrixConfig == nil {
		httpClient := httpclient.NewClient(
			httpclient.WithHTTPClient(getClient(requestConfig)),
			httpclient.WithHTTPTimeout(requestConfig.timeout*time.Millisecond),
			httpclient.WithRetryCount(requestConfig.retryCount),
			httpclient.WithRetrier(getRetrier(requestConfig)),
		)
		return httpClient
	} else {
		hystixClient := hystrix.NewClient(
			hystrix.WithHTTPClient(getClient(requestConfig)),
			hystrix.WithCommandName(requestConfig.name),
			hystrix.WithHTTPTimeout(requestConfig.timeout*time.Millisecond),
			hystrix.WithRetryCount(requestConfig.retryCount),
			hystrix.WithRetrier(getRetrier(requestConfig)),
			hystrix.WithHystrixTimeout(requestConfig.hystrixConfig.hystrixTimeout*time.Millisecond),
			hystrix.WithMaxConcurrentRequests(requestConfig.hystrixConfig.maxConcurrentRequests),
			hystrix.WithErrorPercentThreshold(requestConfig.hystrixConfig.errorPercentThreshold),
			hystrix.WithSleepWindow(int(requestConfig.hystrixConfig.sleepWindow)),
			hystrix.WithRequestVolumeThreshold(10),
			hystrix.WithFallbackFunc(requestConfig.hystrixConfig.fallback),
		)

		return hystixClient
	}

}

// This creates http client and setup transport based on RequestConfig settings.
// Following are default transport settings:
// ForceAttemptHTTP2 : true
// MaxIdleConnsPerHost : runtime.GOMAXPROCS(0) + 1
func getClient(requestConfig *RequestConfig) heimdall.Doer {

	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

	if err != nil {
		log.Fatal(err)
	}

	dialer := &net.Dialer{
		Timeout:   requestConfig.timeout * time.Millisecond,
		KeepAlive: requestConfig.keepalive * time.Millisecond,
	}

	client := &http.Client{
		Jar:     cookieJar,
		Timeout: requestConfig.timeout * time.Millisecond,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          requestConfig.maxIdleConnections,
			IdleConnTimeout:       requestConfig.idleConnectionTimeout,
			TLSHandshakeTimeout:   requestConfig.tlsHandshakeTimeout,
			ExpectContinueTimeout: requestConfig.expectContinueTimeout,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}

	client = setProxy(requestConfig, client)

	return client
}

// This sets the proxy to transport using proxy provided in RequestConfig
func setProxy(requestConfig *RequestConfig, client *http.Client) *http.Client {
	if requestConfig.proxyURL != "" {
		transport, err := transport(client)
		if err != nil {
			log.Printf("%v", err)
			return client
		}

		pURL, err := url.Parse(requestConfig.proxyURL)
		if err != nil {
			log.Printf("%v", err)
			return client
		}

		transport.Proxy = http.ProxyURL(pURL)

	}

	return client

}

// Transport method returns `*http.Transport` currently in use or error
// in case currently used `transport` is not a `*http.Transport`.
func transport(c *http.Client) (*http.Transport, error) {
	if transport, ok := c.Transport.(*http.Transport); ok {
		return transport, nil
	}
	return nil, errors.New("current transport is not an *http.Transport instance")
}

// This constructs the retry function (ConstantBackoff, ExponentialBackoff or NoRetrier) based on
// BackoffPolicy settings provided in RequestConfig
// NoRetry is used if no BackoffPolicy setting are provided
func getRetrier(requestConfig *RequestConfig) heimdall.Retriable {
	if requestConfig.backoffPolicy != nil && requestConfig.backoffPolicy.constantBackoff != nil {
		return heimdall.NewRetrier(heimdall.NewConstantBackoff(requestConfig.backoffPolicy.constantBackoff.interval,
			requestConfig.backoffPolicy.constantBackoff.maximumJitterInterval))
	} else if requestConfig.backoffPolicy != nil && requestConfig.backoffPolicy.exponentialBackoff != nil {
		return heimdall.NewRetrier(heimdall.NewExponentialBackoff(
			requestConfig.backoffPolicy.exponentialBackoff.initialTimeout,
			requestConfig.backoffPolicy.exponentialBackoff.maxTimeout,
			requestConfig.backoffPolicy.exponentialBackoff.exponentFactor,
			requestConfig.backoffPolicy.exponentialBackoff.maximumJitterInterval))
	}

	return heimdall.NewNoRetrier()
}
