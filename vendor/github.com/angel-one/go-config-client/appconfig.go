package configs

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appconfig"
	jsoniter "github.com/json-iterator/go"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/yaml.v3"
	"net"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"strings"
	"sync"
	"time"
)

const defaultAppConfigCheckInterval = time.Minute

var ErrorAppConfigNotAdded = errors.New("config not added")
var ErrAppConfigNotFound = errors.New("config not found")
var ErrAppConfigInvalidType = errors.New("config invalid type")

type appConfigClient struct {
	options   appConfigClientOptions
	session   *session.Session
	client    *appconfig.AppConfig
	parser    func([]byte, interface{}) error
	signal    chan struct{}
	configs   map[string]*appConfig
	mu        sync.RWMutex
	listeners map[string]ChangeListener
}

type appConfig struct {
	mu      sync.RWMutex
	version string
	data    map[string]interface{}
}

type appConfigClientOptions struct {
	id            string
	region        string
	accessKeyID   string
	secretKey     string
	token         string
	app           string
	env           string
	configType    string
	configNames   []string
	checkInterval time.Duration
}

type appConfigClientLogger struct{}

type appConfigHTTPClientConfig struct {
	connectTimeout        time.Duration
	keepAliveDuration     time.Duration
	maxIdleConnections    int
	idleConnectionTimeout time.Duration
	tlsHandshakeTimeout   time.Duration
	expectContinueTimeout time.Duration
	timeout               time.Duration
}

func (*appConfigClientLogger) Log(args ...interface{}) {
	log.Print(args...)
	log.Print("\n")
}

func getAppConfigOption(options map[string]interface{}, key string) (string, error) {
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

func getAppConfigClientOptions(options map[string]interface{}) (appConfigClientOptions, error) {
	var clientOptions appConfigClientOptions
	var err error
	clientOptions.id, err = getAppConfigOption(options, "id")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.region, err = getAppConfigOption(options, "region")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.accessKeyID, err = getAppConfigOption(options, "accessKeyId")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.secretKey, err = getAppConfigOption(options, "secretKey")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.app, err = getAppConfigOption(options, "app")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.env, err = getAppConfigOption(options, "env")
	if err != nil {
		return clientOptions, err
	}
	clientOptions.configType, err = getAppConfigOption(options, "configType")
	if err != nil {
		return clientOptions, err
	}
	if val, ok := options["configNames"]; ok {
		if clientOptions.configNames, ok = val.([]string); !ok {
			return clientOptions, errors.New("invalid config names provided, should be an array of strings")
		}
	} else {
		return clientOptions, errors.New("missing configs")
	}
	if val, ok := options["checkInterval"]; ok {
		if clientOptions.checkInterval, ok = val.(time.Duration); !ok {
			return clientOptions, errors.New("invalid check interval provided, must be a time duration")
		}
	}
	if clientOptions.checkInterval == 0 {
		clientOptions.checkInterval = defaultAppConfigCheckInterval
	}
	return clientOptions, nil
}

func getAppConfigHTTPClientConfig(options map[string]interface{}) appConfigHTTPClientConfig {
	// providing the defaults to the http client config
	config := appConfigHTTPClientConfig{
		connectTimeout:        time.Second * 10,
		keepAliveDuration:     time.Second * 30,
		maxIdleConnections:    100,
		idleConnectionTimeout: time.Second * 90,
		tlsHandshakeTimeout:   time.Second * 10,
		expectContinueTimeout: time.Second,
		timeout:               time.Second * 15,
	}
	// now checking for overrides
	if val, ok := options["connectTimeout"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.connectTimeout = d
		}
	}
	if val, ok := options["keepAliveDuration"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.keepAliveDuration = d
		}
	}
	if val, ok := options["maxIdleConnections"]; ok {
		if d, ok := val.(int); ok {
			config.maxIdleConnections = d
		}
	}
	if val, ok := options["idleConnectionTimeout"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.idleConnectionTimeout = d
		}
	}
	if val, ok := options["tlsHandshakeTimeout"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.tlsHandshakeTimeout = d
		}
	}
	if val, ok := options["expectContinueTimeout"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.expectContinueTimeout = d
		}
	}
	if val, ok := options["timeout"]; ok {
		if d, ok := val.(time.Duration); ok {
			config.timeout = d
		}
	}
	return config
}

func (a *appConfigClient) watchConfig(name string, config *appConfig) {
	ticker := time.NewTicker(a.options.checkInterval)
	for {
		select {
		case <-ticker.C:
			// try to fetch the configurations again
			result, err := a.client.GetConfiguration(&appconfig.GetConfigurationInput{
				Application:                aws.String(a.options.app),
				ClientConfigurationVersion: aws.String(config.version),
				ClientId:                   aws.String(a.options.id),
				Configuration:              aws.String(name),
				Environment:                aws.String(a.options.env),
			})
			if err != nil {
				// it might be possible that the configuration is deleted
				// stop the watch
				return
			}
			config.mu.Lock()
			if *result.ConfigurationVersion != config.version {
				// something has changed
				var data map[string]interface{}
				err = a.parser(result.Content, &data)
				if err != nil {
					// someone has added incorrect configurations
					// ignore the change for now then
				} else {
					config.version = *result.ConfigurationVersion
					config.data = data
					// now we also need to notify the listener if any
					a.mu.RLock()
					if l, ok := a.listeners[name]; ok {
						l(config.version, config.data)
					}
					a.mu.RUnlock()
				}
			}
			config.mu.Unlock()
		case <-a.signal:
			// close the watch
			return
		}
	}
}

func (a *appConfigClient) fetchAndWatchConfigs() {
	id := aws.String(a.options.id)
	app := aws.String(a.options.app)
	env := aws.String(a.options.env)
	for _, c := range a.options.configNames {
		result, err := a.client.GetConfiguration(&appconfig.GetConfigurationInput{
			Application:   app,
			ClientId:      id,
			Configuration: aws.String(c),
			Environment:   env,
		})
		if err == nil {
			// now it means that the result exists
			// try to parse the result now
			var data map[string]interface{}
			err = a.parser(result.Content, &data)
			if err == nil {
				config := appConfig{
					version: *result.ConfigurationVersion,
					data:    data,
				}
				a.configs[c] = &config
				go a.watchConfig(c, &config)
			}
		}
	}
}

func getParser(configType string) func([]byte, interface{}) error {
	switch configType {
	case jsonType:
		return jsoniter.Unmarshal
	case yamlType:
		return yaml.Unmarshal
	case tomlType:
		return toml.Unmarshal
	}
	return nil
}

func getAppConfigHTTPClient(options map[string]interface{}) *http.Client {
	if val, ok := options["httpClient"]; ok {
		if c, ok := val.(*http.Client); ok {
			return c
		}
	}
	c := getAppConfigHTTPClientConfig(options)
	cookieJar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return http.DefaultClient
	}
	dialer := &net.Dialer{
		Timeout:   c.connectTimeout,
		KeepAlive: c.keepAliveDuration,
	}

	return &http.Client{
		Jar: cookieJar,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialer.DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          c.maxIdleConnections,
			IdleConnTimeout:       c.idleConnectionTimeout,
			TLSHandshakeTimeout:   c.tlsHandshakeTimeout,
			ExpectContinueTimeout: c.expectContinueTimeout,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
		// global timeout value for all requests
		Timeout: c.timeout,
	}
}

func newAppConfigClient(options map[string]interface{}) (*appConfigClient, error) {
	clientOptions, err := getAppConfigClientOptions(options)
	if err != nil {
		return nil, err
	}
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(clientOptions.accessKeyID,
			clientOptions.secretKey, clientOptions.token),
		Region:     aws.String(clientOptions.region),
		LogLevel:   aws.LogLevel(aws.LogDebug),
		Logger:     &appConfigClientLogger{},
		HTTPClient: getAppConfigHTTPClient(options),
	})
	if err != nil {
		return nil, err
	}
	client := &appConfigClient{
		options:   clientOptions,
		session:   s,
		client:    appconfig.New(s),
		parser:    getParser(clientOptions.configType),
		signal:    make(chan struct{}, 1),
		configs:   make(map[string]*appConfig),
		listeners: make(map[string]ChangeListener),
	}
	client.fetchAndWatchConfigs()
	return client, nil
}

func (a *appConfigClient) AddChangeListener(config string, listener ChangeListener) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.configs[config]; !ok {
		return ErrorAppConfigNotAdded
	}
	a.listeners[config] = listener
	return nil
}

func (a *appConfigClient) RemoveChangeListener(config string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.configs[config]; !ok {
		return ErrorAppConfigNotAdded
	}
	delete(a.listeners, config)
	return nil
}

func get(kList []string, key string, val interface{}) interface{} {
	if len(kList) == 0 {
		if key == "" || key == "." {
			return val
		}
		return nil
	}
	// now need a map
	if m, ok := val.(map[string]interface{}); ok {
		newKey := key + kList[0]
		// for example a.b.c.d, then first try with a
		if v, ok := m[newKey]; ok {
			result := get(kList[1:], "", v)
			if result != nil {
				return result
			}
		}
		// otherwise, proceed with a.[something]
		return get(kList[1:], newKey+".", m)
	}
	// not a map
	return nil
}

func (a *appConfigClient) Get(config, key string) (interface{}, error) {
	if result, ok := a.configs[config]; ok {
		result.mu.RLock()
		defer result.mu.RUnlock()
		if key == "" {
			return result.data, nil
		}
		val := get(strings.Split(key, "."), "", result.data)
		if val == nil {
			return nil, ErrAppConfigNotFound
		}
		return val, nil
	}
	return nil, ErrorAppConfigNotAdded
}

func (a *appConfigClient) GetD(config, key string, defaultValue interface{}) interface{} {
	val, err := a.Get(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetInt(config, key string) (int64, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return 0, err
	}
	if f, ok := val.(float64); ok {
		return int64(f), nil
	}
	return 0, ErrAppConfigInvalidType
}

func (a *appConfigClient) GetIntD(config, key string, defaultValue int64) int64 {
	val, err := a.GetInt(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetFloat(config, key string) (float64, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return 0, err
	}
	if f, ok := val.(float64); ok {
		return f, nil
	}
	return 0, ErrAppConfigInvalidType
}

func (a *appConfigClient) GetFloatD(config, key string, defaultValue float64) float64 {
	val, err := a.GetFloat(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetString(config, key string) (string, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return "", err
	}
	if s, ok := val.(string); ok {
		return s, nil
	}
	return "", ErrAppConfigInvalidType
}

func (a *appConfigClient) GetStringD(config, key string, defaultValue string) string {
	val, err := a.GetString(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetBool(config, key string) (bool, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return false, err
	}
	if b, ok := val.(bool); ok {
		return b, nil
	}
	return false, ErrAppConfigInvalidType
}

func (a *appConfigClient) GetBoolD(config, key string, defaultValue bool) bool {
	val, err := a.GetBool(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetSlice(config, key string) ([]interface{}, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return nil, err
	}
	if sl, ok := val.([]interface{}); ok {
		return sl, nil
	}
	return nil, ErrAppConfigInvalidType
}

func (a *appConfigClient) GetSliceD(config, key string, defaultValue []interface{}) []interface{} {
	val, err := a.GetSlice(config, key)
	if err == nil {
		return val
	}
	return defaultValue
}

func (a *appConfigClient) GetIntSlice(config, key string) ([]int64, error) {
	val, err := a.GetSlice(config, key)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(val))
	for _, v := range val {
		if i, ok := v.(float64); ok {
			res = append(res, int64(i))
		}
	}
	return res, nil
}

func (a *appConfigClient) GetIntSliceD(config, key string, defaultValue []int64) []int64 {
	val, err := a.GetIntSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetFloatSlice(config, key string) ([]float64, error) {
	val, err := a.GetSlice(config, key)
	if err != nil {
		return nil, err
	}
	res := make([]float64, 0, len(val))
	for _, v := range val {
		if f, ok := v.(float64); ok {
			res = append(res, f)
		}
	}
	return res, nil
}

func (a *appConfigClient) GetFloatSliceD(config, key string, defaultValue []float64) []float64 {
	val, err := a.GetFloatSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetStringSlice(config, key string) ([]string, error) {
	val, err := a.GetSlice(config, key)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0, len(val))
	for _, v := range val {
		if s, ok := v.(string); ok {
			res = append(res, s)
		}
	}
	return res, nil
}

func (a *appConfigClient) GetStringSliceD(config, key string, defaultValue []string) []string {
	val, err := a.GetStringSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetBoolSlice(config, key string) ([]bool, error) {
	val, err := a.GetSlice(config, key)
	if err != nil {
		return nil, err
	}
	res := make([]bool, 0, len(val))
	for _, v := range val {
		if b, ok := v.(bool); ok {
			res = append(res, b)
		}
	}
	return res, nil
}

func (a *appConfigClient) GetBoolSliceD(config, key string, defaultValue []bool) []bool {
	val, err := a.GetBoolSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetMap(config, key string) (map[string]interface{}, error) {
	val, err := a.Get(config, key)
	if err != nil {
		return nil, err
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, ErrAppConfigInvalidType
}

func (a *appConfigClient) GetMapD(config, key string, defaultValue map[string]interface{}) map[string]interface{} {
	val, err := a.GetMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetIntMap(config, key string) (map[string]int64, error) {
	val, err := a.GetMap(config, key)
	if err != nil {
		return nil, err
	}
	res := make(map[string]int64)
	for k, v := range val {
		if i, ok := v.(float64); ok {
			res[k] = int64(i)
		}
	}
	return res, nil
}

func (a *appConfigClient) GetIntMapD(config, key string, defaultValue map[string]int64) map[string]int64 {
	val, err := a.GetIntMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetFloatMap(config, key string) (map[string]float64, error) {
	val, err := a.GetMap(config, key)
	if err != nil {
		return nil, err
	}
	res := make(map[string]float64)
	for k, v := range val {
		if f, ok := v.(float64); ok {
			res[k] = f
		}
	}
	return res, nil
}

func (a *appConfigClient) GetFloatMapD(config, key string, defaultValue map[string]float64) map[string]float64 {
	val, err := a.GetFloatMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetStringMap(config, key string) (map[string]string, error) {
	val, err := a.GetMap(config, key)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for k, v := range val {
		if s, ok := v.(string); ok {
			res[k] = s
		}
	}
	return res, nil
}

func (a *appConfigClient) GetStringMapD(config, key string, defaultValue map[string]string) map[string]string {
	val, err := a.GetStringMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) GetBoolMap(config, key string) (map[string]bool, error) {
	val, err := a.GetMap(config, key)
	if err != nil {
		return nil, err
	}
	res := make(map[string]bool)
	for k, v := range val {
		if b, ok := v.(bool); ok {
			res[k] = b
		}
	}
	return res, nil
}

func (a *appConfigClient) GetBoolMapD(config, key string, defaultValue map[string]bool) map[string]bool {
	val, err := a.GetBoolMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (a *appConfigClient) Close() error {
	a.signal <- struct{}{}
	return nil
}
