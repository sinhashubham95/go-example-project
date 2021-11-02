package httpclient

import (
	httpclient "github.com/angel-one/go-http-client"
	"net/http"
)

// Client is the set of methods for the http client
type Client interface {
	Request(request *httpclient.Request) (*http.Response, error)
}

var client *httpclient.Client

// Init is used to initialise the http client
func Init(configs ...*httpclient.RequestConfig) {
	client = httpclient.ConfigureHTTPClient(configs...)
}

// NewRequestConfig is used to create a new request config
func NewRequestConfig(name string, configs map[string]interface{}) *httpclient.RequestConfig {
	return httpclient.NewRequestConfig(name, configs)
}

// NewRequest is used to create a new request
func NewRequest(name string) *httpclient.Request {
	return httpclient.NewRequest(name)
}

// Get is used to get the client instance
func Get() Client {
	return client
}
