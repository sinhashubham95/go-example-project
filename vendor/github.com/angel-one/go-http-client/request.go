package httpclient

import (
	"golang.org/x/net/context"
	"io"
)

// NewRequest creates a Request to execute
// It takes request name as input. It must match with request name used to configure RequestConfig
func NewRequest(name string) *Request {
	//add defaults
	r := &Request{
		name: name,
	}
	return r
}

// Request is the type for the request created
type Request struct {
	name         string
	ctx          context.Context
	url          string
	queryParams  map[string]string
	headerParams map[string]string
	body         io.Reader
}

// SetContext is used to set the context for the request
func (req *Request) SetContext(ctx context.Context) *Request {
	req.ctx = ctx
	return req
}

// SetURL is used to set the url for the request
// if not done, then the url already configured will be used
func (req *Request) SetURL(url string) *Request {
	req.url = url
	return req
}

// SetQueryParam is used to set a query param key value pair
// These will be passed in query param while executing HTTP request
func (req *Request) SetQueryParam(param, value string) *Request {
	if req.queryParams == nil {
		req.queryParams = make(map[string]string)
	}
	req.queryParams[param] = value
	return req
}

// SetQueryParams is used to set multiple query params - map of key-value pair
// These will be passed in query param while executing HTTP request
func (req *Request) SetQueryParams(queryParams map[string]string) *Request {
	for p, v := range queryParams {
		req.SetQueryParam(p, v)
	}
	return req
}

// SetHeaderParam is used to set a header - key-value pair
// These will be passed in header while executing HTTP request
func (req *Request) SetHeaderParam(param, value string) *Request {
	if req.headerParams == nil {
		req.headerParams = make(map[string]string)
	}
	req.headerParams[param] = value
	return req
}

// SetHeaderParams is used to set multiple headers -  map of key-value pair
// These will be passed in header while executing HTTP request
func (req *Request) SetHeaderParams(headerParams map[string]string) *Request {
	req.headerParams = headerParams
	return req
}

// SetBody is used to set request body to pass in http request
func (req *Request) SetBody(body io.Reader) *Request {
	req.body = body
	return req
}
