package httpclient

import (
	"fmt"
	"time"
)

// RequestConfig is the type for a request configuration
type RequestConfig struct {
	name                  string
	method                string
	url                   string
	timeout               time.Duration
	keepalive             time.Duration
	maxIdleConnections    int
	idleConnectionTimeout time.Duration
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
	proxyURL              string
	retryCount            int
	backoffPolicy         *BackoffPolicy
	hystrixConfig         *HystrixConfig
}

// NewRequestConfig is used to create a new request configuration from a map of configurations.
func NewRequestConfig(name string, configMap map[string]interface{}) *RequestConfig {
	rc := RequestConfig{
		name: name,
	}

	if configMap != nil {
		var err error

		rc.method, err = getConfigOptionString(configMap, "method")
		if err != nil {
			return &rc
		}

		rc.url, err = getConfigOptionString(configMap, "url")
		if err != nil {
			return &rc
		}

		timeout, err := getConfigOptionInt(configMap, "timeoutinmillis")
		if err != nil {
			return &rc
		}
		rc.timeout = time.Duration(timeout) * time.Millisecond

		keepalive, err := getConfigOptionInt(configMap, "keepaliveinmillis")
		if err == nil {
			rc.keepalive = time.Duration(keepalive) * time.Millisecond
		}

		rc.maxIdleConnections, _ = getConfigOptionInt(configMap, "maxidleonnections")

		idleConnectionTimeout, err := getConfigOptionInt(configMap, "idleconnectiontimeoutinmillis")
		if err == nil {
			rc.idleConnectionTimeout = time.Duration(idleConnectionTimeout) * time.Millisecond
		}

		tlsHandshakeTimeout, err := getConfigOptionInt(configMap, "tlshandshaketimeoutinmillis")
		if err == nil {
			rc.tlsHandshakeTimeout = time.Duration(tlsHandshakeTimeout) * time.Millisecond
		}

		expectContinueTimeout, err := getConfigOptionInt(configMap, "expectcontinuetimeoutinmillis")
		if err == nil {
			rc.expectContinueTimeout = time.Duration(expectContinueTimeout) * time.Millisecond
		}

		rc.proxyURL, _ = getConfigOptionString(configMap, "proxyurl")

		rc.retryCount, err = getConfigOptionInt(configMap, "retrycount")
		if err != nil {
			rc.retryCount = 1
		}

		backoffPolicyMap, err := getConfigOptionMap(configMap, "backoffpolicy")
		if err == nil {
			rc.backoffPolicy = NewBackoffPolicy(backoffPolicyMap)
		}

		hystrixConfig, err := getConfigOptionMap(configMap, "hystrixconfig")
		if err == nil {
			rc.hystrixConfig = NewHystrixConfig(hystrixConfig)
		}

	}
	return &rc
}

// SetName is used to set name for request
func (rc *RequestConfig) SetName(name string) *RequestConfig {
	rc.name = name
	return rc
}

// SetMethod is used to set method for request
func (rc *RequestConfig) SetMethod(method string) *RequestConfig {
	rc.method = method
	return rc
}

// SetURL is used to set the url for request
func (rc *RequestConfig) SetURL(url string) *RequestConfig {
	rc.url = url
	return rc
}

// SetProxy is used to set the proxy url for request
func (rc *RequestConfig) SetProxy(proxyURL string) *RequestConfig {
	rc.proxyURL = proxyURL
	return rc
}

// SetTimeout is used to set the timeout for request
func (rc *RequestConfig) SetTimeout(timeout time.Duration) *RequestConfig {
	rc.timeout = timeout
	return rc
}

// SetKeepAlive is used to set the keep alive for request
func (rc *RequestConfig) SetKeepAlive(keepalive time.Duration) *RequestConfig {
	rc.keepalive = keepalive
	return rc
}

// SetMaxIdleConnections is used to set the max idle connections for request
func (rc *RequestConfig) SetMaxIdleConnections(maxIdleConnections int) *RequestConfig {
	rc.maxIdleConnections = maxIdleConnections
	return rc
}

// SetIdleConnectionTimeout is used to set the idle connection timeout for request
func (rc *RequestConfig) SetIdleConnectionTimeout(idleConnectionTimeout time.Duration) *RequestConfig {
	rc.idleConnectionTimeout = idleConnectionTimeout
	return rc
}

// SetTLSHandshakeTimeout is used to set the tls handshake timeout for request
func (rc *RequestConfig) SetTLSHandshakeTimeout(tlsHandshakeTimeout time.Duration) *RequestConfig {
	rc.tlsHandshakeTimeout = tlsHandshakeTimeout
	return rc
}

// SetExpectContinueTimeout is used to set expect continue timeout for request
func (rc *RequestConfig) SetExpectContinueTimeout(expectContinueTimeout time.Duration) *RequestConfig {
	rc.expectContinueTimeout = expectContinueTimeout
	return rc
}

// SetRetryCount is used to set the retry count for request
func (rc *RequestConfig) SetRetryCount(retryCount int) *RequestConfig {
	rc.retryCount = retryCount
	return rc
}

// SetBackoffPolicy is used to set the backoff policy for request
func (rc *RequestConfig) SetBackoffPolicy(backoffPolicy *BackoffPolicy) *RequestConfig {
	rc.backoffPolicy = backoffPolicy
	return rc
}

// SetHystrixConfig is used to set the hystrix config for the request
func (rc *RequestConfig) SetHystrixConfig(hystrixConfig *HystrixConfig) *RequestConfig {
	rc.hystrixConfig = hystrixConfig
	return rc
}

// SetHystrixFallback is used to set the hystrix fallback for the request
func (rc *RequestConfig) SetHystrixFallback(fallbackFn func(error) error) *RequestConfig {
	if rc.hystrixConfig != nil {
		rc.hystrixConfig.fallback = fallbackFn
	}
	return rc
}

func getConfigOptionInt(options map[string]interface{}, key string) (int, error) {
	var val interface{}
	var ok bool
	var s int
	if val, ok = options[key]; ok {
		if s, ok = val.(int); !ok {
			return s, fmt.Errorf("invalid %s, must be a int", key)
		}
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
	return s, nil
}

func getConfigOptionFloat(options map[string]interface{}, key string) (float64, error) {
	var val interface{}
	var ok bool
	var s float64
	if val, ok = options[key]; ok {
		if s, ok = val.(float64); !ok {
			return s, fmt.Errorf("invalid %s, must be a float64", key)
		}
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
	return s, nil
}

func getConfigOptionMap(options map[string]interface{}, key string) (map[string]interface{}, error) {
	var val interface{}
	var ok bool
	var s map[string]interface{}
	if val, ok = options[key]; ok {
		if s, ok = val.(map[string]interface{}); !ok {
			return s, fmt.Errorf("invalid %s, must be a map[string]interface{}", key)
		}
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
	return s, nil
}

func getConfigOptionString(options map[string]interface{}, key string) (string, error) {
	var val interface{}
	var ok bool
	var s string
	if val, ok = options[key]; ok {
		if s, ok = val.(string); !ok {
			return s, fmt.Errorf("invalid %s, must be a string", key)
		}
	} else {
		return s, fmt.Errorf("missing %s", key)
	}
	return s, nil
}
