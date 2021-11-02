# Go Config Client

This is the config client for Go projects.

- **File-Based Client** - Fully Implemented
- **ETCD Client** - TBD
- **AWS App Config Client** - Fully Implemented

## Project Versioning

Go Config Client uses [semantic versioning](http://semver.org/). API should not change between patch and minor releases. New minor versions may add additional features to the API.

## Installation

To install `Go Config Client` package, you need to install Go and set your Go workspace first.

1. The first need Go installed (version 1.13+ is required), then you can use the below Go command to install Go Config Client.

```shell
go get commit.angelbroking.com/foundation/go-config-client
```

2. Because this is a private repository, you will need to mark this in the Go env variables.

```shell
go env -w GOPRIVATE=github.com/angel-one/go-utils
```

3. Also, follow this to generate a personal access token and add the following line to your $HOME/.netrc file.

```
machine github.com login ${USERNAME} password ${PERSONAL_ACCESS_TOKEN}
```

4. Import it in your code:

```go
import configs "commit.angelbroking.com/foundation/go-config-client"
```

## Usage

### New Client

```go
import configs "commit.angelbroking.com/foundation/go-config-client"

fileBasedClient, err := configs.New(configs.Options{
    Provider: configs.FileBased,
    Params: map[string]interface{}{
        "configsDirectory": ".",
        "configNames": []string{"configs"},
        "configType": "json",
    },
})
if err != nil {
	// handle error
}

awsAppConfigClient, err := configs.New(configs.Options{
	Provider: configs.AWSAppConfig,
	Params: map[string]interface{}{
        "id":          "example-go-config-client",
        "region":      os.Getenv("region"),
        "accessKeyId": os.Getenv("accessKeyId"),
        "secretKey":   os.Getenv("secretKey"),
        "app":         "sample",
        "env":         "sample",
        "configType":  "json",
        "configNames": []string{"sample"},
    }
})
```

### Getting Configs

There are 2 types of methods available.
1. Plain methods which take the config name and the key.
2. Methods with default values which take the config name, key and the default value. The default value will be used in case the value is not found in the config mentioned corresponding to the key asked for.

For **File Based Config Client**, the config name is the name of the file from where the configurations have to be referenced, and the key is the location of the config being fetched from that configuration file.

For **AWS App Config Client**, the config name is the name of the configuration profile deployed, and the key is the location of the config being fetched from that configuration profile.
