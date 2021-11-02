package configs

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"path/filepath"
	"sync"
)

var ErrorFileBasedConfigNotAdded = errors.New("config not added")

type fileBasedClient struct {
	options   fileBasedClientOptions
	configs   map[string]*viper.Viper
	listeners map[string]ChangeListener
	mu        sync.RWMutex
}

type fileBasedClientOptions struct {
	path        string
	configNames []string
	configType  string
}

func getFileBasedClientOptions(options map[string]interface{}) (fileBasedClientOptions, error) {
	clientOptions := fileBasedClientOptions{}
	var val interface{}
	var ok bool
	if val, ok = options["configsDirectory"]; ok {
		if clientOptions.path, ok = val.(string); ok {
			clientOptions.path = filepath.Clean(clientOptions.path)
		} else {
			return clientOptions, errors.New("invalid config directory provided")
		}
	} else {
		return clientOptions, errors.New("config directory not provided")
	}
	if val, ok = options["configNames"]; ok {
		if clientOptions.configNames, ok = val.([]string); !ok {
			return clientOptions, errors.New("invalid config names provided, should be an array of strings")
		}

	} else {
		return clientOptions, errors.New("no configs provided to be used")
	}
	if val, ok = options["configType"]; ok {
		if clientOptions.configType, ok = val.(string); !ok || (clientOptions.configType != jsonType &&
			clientOptions.configType != yamlType &&
			clientOptions.configType != tomlType) {
			return clientOptions, fmt.Errorf("invalid config type provided should be one of %s, %s or %s",
				jsonType, yamlType, tomlType)
		}
	} else {
		return clientOptions, errors.New("no config type provided")
	}
	return clientOptions, nil
}

func (f *fileBasedClient) onConfigChange(e fsnotify.Event) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if l, ok := f.listeners[e.Name]; ok {
		l(e.Name, e.Op)
	}
}

func (f *fileBasedClient) getConfig(options fileBasedClientOptions, name string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType(options.configType)
	v.AddConfigPath(options.path)
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	v.OnConfigChange(f.onConfigChange)
	v.WatchConfig()
	return v, nil
}

func newFileBasedClient(options map[string]interface{}) (*fileBasedClient, error) {
	clientOptions, err := getFileBasedClientOptions(options)
	if err != nil {
		return nil, err
	}
	client := &fileBasedClient{
		options: clientOptions,
	}
	client.configs = make(map[string]*viper.Viper)
	for _, name := range clientOptions.configNames {
		v, err := client.getConfig(clientOptions, name)
		if err != nil {
			return nil, err
		}
		client.configs[name] = v
	}
	client.listeners = make(map[string]ChangeListener)
	return client, nil
}

func (f *fileBasedClient) AddChangeListener(config string, listener ChangeListener) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listeners[config] = listener
	return nil
}

func (f *fileBasedClient) RemoveChangeListener(config string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.listeners, config)
	return nil
}

func (f *fileBasedClient) Get(config, key string) (interface{}, error) {
	if v, ok := f.configs[config]; ok {
		return v.Get(key), nil
	} else {
		return nil, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetD(config, key string, defaultValue interface{}) interface{} {
	val, err := f.Get(config, key)
	if err != nil || val == nil {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetInt(config, key string) (int64, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetInt64(key), nil
	} else {
		return 0, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetIntD(config, key string, defaultValue int64) int64 {
	val, err := f.GetInt(config, key)
	if err != nil || val == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetFloat(config, key string) (float64, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetFloat64(key), nil
	} else {
		return 0, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetFloatD(config, key string, defaultValue float64) float64 {
	val, err := f.GetFloat(config, key)
	if err != nil || val == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetString(config, key string) (string, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetString(key), nil
	} else {
		return "", ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetStringD(config, key string, defaultValue string) string {
	val, err := f.GetString(config, key)
	if err != nil || val == "" {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetBool(config, key string) (bool, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetBool(key), nil
	} else {
		return false, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetBoolD(config, key string, defaultValue bool) bool {
	val, err := f.GetBool(config, key)
	if err != nil || val == false {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetSlice(config, key string) ([]interface{}, error) {
	val, err := f.Get(config, key)
	if err != nil {
		return nil, err
	}
	if a, ok := val.([]interface{}); ok {
		return a, nil
	}
	return nil, errors.New("invalid value type")
}

func (f *fileBasedClient) GetSliceD(config, key string, defaultValue []interface{}) []interface{} {
	val, err := f.GetSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetIntSlice(config, key string) ([]int64, error) {
	val, err := f.GetSlice(config, key)
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

func (f *fileBasedClient) GetIntSliceD(config, key string, defaultValue []int64) []int64 {
	val, err := f.GetIntSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetFloatSlice(config, key string) ([]float64, error) {
	val, err := f.GetSlice(config, key)
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

func (f *fileBasedClient) GetFloatSliceD(config, key string, defaultValue []float64) []float64 {
	val, err := f.GetFloatSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetStringSlice(config, key string) ([]string, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetStringSlice(key), nil
	} else {
		return nil, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetStringSliceD(config, key string, defaultValue []string) []string {
	val, err := f.GetStringSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetBoolSlice(config, key string) ([]bool, error) {
	val, err := f.GetSlice(config, key)
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

func (f *fileBasedClient) GetBoolSliceD(config, key string, defaultValue []bool) []bool {
	val, err := f.GetBoolSlice(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetMap(config, key string) (map[string]interface{}, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetStringMap(key), nil
	} else {
		return nil, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetMapD(config, key string, defaultValue map[string]interface{}) map[string]interface{} {
	val, err := f.GetMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetIntMap(config, key string) (map[string]int64, error) {
	val, err := f.GetMap(config, key)
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

func (f *fileBasedClient) GetIntMapD(config, key string, defaultValue map[string]int64) map[string]int64 {
	val, err := f.GetIntMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetFloatMap(config, key string) (map[string]float64, error) {
	val, err := f.GetMap(config, key)
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

func (f *fileBasedClient) GetFloatMapD(config, key string, defaultValue map[string]float64) map[string]float64 {
	val, err := f.GetFloatMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetStringMap(config, key string) (map[string]string, error) {
	if v, ok := f.configs[config]; ok {
		return v.GetStringMapString(key), nil
	} else {
		return nil, ErrorFileBasedConfigNotAdded
	}
}

func (f *fileBasedClient) GetStringMapD(config, key string, defaultValue map[string]string) map[string]string {
	val, err := f.GetStringMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) GetBoolMap(config, key string) (map[string]bool, error) {
	val, err := f.GetMap(config, key)
	if err != nil {
		return nil, err
	}
	res := make(map[string]bool)
	for k, v := range val {
		if f, ok := v.(bool); ok {
			res[k] = f
		}
	}
	return res, nil
}

func (f *fileBasedClient) GetBoolMapD(config, key string, defaultValue map[string]bool) map[string]bool {
	val, err := f.GetBoolMap(config, key)
	if err != nil || len(val) == 0 {
		return defaultValue
	}
	return val
}

func (f *fileBasedClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for k := range f.configs {
		delete(f.configs, k)
	}
	for k := range f.listeners {
		delete(f.listeners, k)
	}
	return nil
}
