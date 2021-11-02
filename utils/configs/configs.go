package configs

import (
	configs "github.com/angel-one/go-config-client"
)

var client configs.Client

// Init is used to initialize the configs
func Init(directory string, configNames ...string) error {
	var err error
	client, err = configs.New(configs.Options{
		Provider: configs.FileBased,
		Params: map[string]interface{}{
			"configsDirectory": directory,
			"configNames":      configNames,
			"configType":       "yaml",
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// Get is used to get the config client
func Get() configs.Client {
	return client
}
