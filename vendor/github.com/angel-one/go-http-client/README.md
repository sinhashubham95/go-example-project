# go-http-client

# Installation

```
go env -w GOPRIVATE=github.com/angel-one/go-http-client
```

```
go get https://github.com/angel-one/go-http-client
```


## Initialisation

### Import module

```
import github.com/angel-one/go-http-client
```



### Initialize go-http-client

#### Using API
Application must configure http client before making any http call. All endpoint must be configured.

For each http endpoint create a NewRequestConfig. NewRequestConfig provides following tunables

|field/method| description | optional|
|----|--------------|--------|
|unique request name|  A unique name must be passed as parameter to NewRequestConfig. In case of duplicate, latest will replace the previous configuration|mandatory|
|SetTimeout| Http timeout| mandatory|
|SetRetryCount| Retry count| mandatory|
|SetMethod| Http Method (GET, POST etc)| mandatory|
|SetURL| Endpoint to call| mandatory|
|SetProxy| Proxy URL |optional|
|SetBackoffPolicy| Backoff policy - you can choose between ConstantBackoff or ExponentialBackoff|optional for NoBackoff|
|SetHystrixConfig| Hystrix Configuration|optional|
|keepalive   |  KeepAliveDuration specifies the interval between keep-alive probes for an active network connection         |mandatory|
|maxIdleConnections | MaxIdleConnections controls the maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.  |mandatory|
|idleConnectionTimeout| IdleConnectionTimeout is the maximum amount of time an idle (keep-alive) connection will remain idle before closing itself.|mandatory|
|tlsHandshakeTimeout  | TLSHandshakeTimeout specifies the maximum amount of time waiting to wait for a TLS handshake.|mandatory|
|expectContinueTimeout| ExpectContinueTimeout specifies the amount of time to wait for a server's first response headers after fully writing the request headers.|mandatory|


```
requestConfig := NewRequestConfig("test", nil).SetTimeout(1000).
		SetRetryCount(3).
		SetMethod("GET").SetURL("http://google.com")
```

#### Using config map

Applications can use yaml files to configure all the http configurations. 
Use https://github.com/angel-one/go-config-client to read yaml and get map[string]interface{}

```
// config map - note that constant backoff is given more preference over exponential backoff if both are set
configMap := map[string]interface{}{
    "method":          "GET",
    "url":             "https://www.google.co.in",
    "timeoutinmillis": 5000,
    "retrycount":      3,
    "backoffpolicy": map[string]interface{}{
        "constantbackoff": map[string]interface{}{
            "intervalinmillis":          2,
            "maxjitterintervalinmillis": 5,
        },
        "exponentialbackoff": map[string]interface{}{
            "initialtimeoutinmillis":    2,
            "maxtimeoutinmillis":        10,
            "exponentfactor":            2.0,
            "maxjitterintervalinmillis": 2,
        },
    },
    "hystrixconfig": map[string]interface{}{
        "maxconcurrentrequests":  10,
        "errorpercentthreshold":  20,
        "sleepwindowinmillis":    10,
        "requestvolumethreshold": 10,
    },
}

//setup activity
requestConfig := NewRequestConfig("test", configMap)
```

#### Configure Client using NewRequestConfig
You can pass as many requestConfig
```
httpclient := ConfigureHTTPClient(*requestConfig)
```
This can also be used to reconfigure client for exist request.

#### Making a request

```
res, error := httpclient.Request(
		NewRequest("test"),
	)
```
NewRequest has following tunables

|field/method| description | optional|
|----|--------------|--------|
|request name|  A unique name must be passed as parameter|mandatory|
|SetQueryParam| set a query param| optional|
|SetQueryParams| set map query params| optional|
|SetHeaderParams| set headers| optional|
|SetBody| set request body| optional|

#### Yaml config

Keys must match NewRequestConfig struct

```yaml
sample-call-1:
  method: "GET"
  url: "http://google.com"
  timeoutInMillis: 1000
  retryCount: 3
  backoffPolicy:
    constantBackoff:
      intervalInMillis: 2
      maxJitterIntervalInMillis: 5
    exponentialBackoff:
      initialTimeoutInMillis: 2
      maxTimeoutInMillis: 10
      exponentFactor: 2
      maxJitterIntervalInMillis: 2
  hystrixConfig:
    maxConcurrentRequests: 10
    errorPercentThresold: 20
    sleepWindowInMillis : 10
    requestVolumeThreshold: 10

sample-call-2:
  method: "GET"
  url: "http://google.com"
  timeoutInMillis: 1000
  retryCount: 3
  backoffPolicy:
    constantBackoff:
      intervalInMillis: 2
      maxJitterIntervalInMillis: 5
    exponentialBackoff:
      initialTimeoutInMillis: 2
      maxTimeoutInMillis: 10
      exponentFactor: 2
      maxJitterIntervalInMillis: 2
  hystrixConfig:
    maxConcurrentRequests: 10
    errorPercentThresold: 20
    sleepWindowInMillis : 10
    requestVolumeThreshold: 10
```
