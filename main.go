package main

import (
	"fmt"
	"github.com/sinhashubham95/go-example-project/api"
	"github.com/sinhashubham95/go-example-project/constants"
	"github.com/sinhashubham95/go-example-project/utils/configs"
	"github.com/sinhashubham95/go-example-project/utils/database"
	"github.com/sinhashubham95/go-example-project/utils/flags"
	"github.com/sinhashubham95/go-example-project/utils/httpclient"
	"github.com/angel-one/go-utils/log"
	"github.com/angel-one/go-utils/middlewares"
	"time"

	_ "github.com/sinhashubham95/go-example-project/docs"
	_ "github.com/go-sql-driver/mysql"
)

// @title Go Example Project
// @version 1.0
// @description Go Example Project
// @termsOfService https://swagger.io/terms/

// @contact.name Shubham Sinha
// @contact.email shubham.sinha@angelbroking.com

// @BasePath /

func main() {
	initConfigs()
	startLogger()
	initHTTPClient()
	initDatabase()
	defer closeDatabase()
	startRouter()
}

func initConfigs() {
	// init configs
	configs.Init(flags.BaseConfigPath())
}

func startLogger() {
	// start logger
	loggerConfig, err := configs.Get(constants.LoggerConfig)
	if err != nil {
		log.Fatal(nil).Err(err).Msg("error getting logger config")
	}
	log.InitLogger(log.Level(loggerConfig.GetString(constants.LogLevelConfigKey)))
}

func initHTTPClient() {
	// get application configs
	applicationConfig, err := configs.Get(constants.ApplicationConfig)
	if err != nil {
		log.Fatal(nil).Err(err).Msg("error getting application config")
	}

	// init http client
	err = httpclient.Init(httpclient.Config{
		ConnectTimeout: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPConnectTimeoutInMillisKey),
		KeepAliveDuration: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPKeepAliveDurationInMillisKey),
		MaxIdleConnections: applicationConfig.GetInt(constants.HTTPMaxIdleConnectionsKey),
		IdleConnectionTimeout: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPIdleConnectionTimeoutInMillisKey),
		TLSHandshakeTimeout: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPTlsHandshakeTimeoutInMillisKey),
		ExpectContinueTimeout: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPExpectContinueTimeoutInMillisKey),
		Timeout: time.Millisecond *
			applicationConfig.GetDuration(constants.HTTPTimeoutInMillisKey),
	})
	if err != nil {
		log.Fatal(nil).Err(err).Msg("unable to initialize http client")
	}
}

func initDatabase() {
	// init database
	databaseConfig, err := configs.Get(constants.DatabaseConfig)
	if err != nil {
		log.Fatal(nil).Err(err).Msg("error getting database config")
	}
	err = database.InitDatabase(database.Config{
		Server:                databaseConfig.GetString(constants.DatabaseServerConfigKey),
		Port:                  databaseConfig.GetInt(constants.DatabasePortConfigKey),
		Name:                  databaseConfig.GetString(constants.DatabaseNameConfigKey),
		Username:              databaseConfig.GetString(constants.DatabaseUsernameConfigKey),
		Password:              databaseConfig.GetString(constants.DatabasePasswordConfigKey),
		MaxOpenConnections:    databaseConfig.GetInt(constants.DatabaseMaxOpenConnectionsKey),
		MaxIdleConnections:    databaseConfig.GetInt(constants.DatabaseMaxIdleConnectionsKey),
		ConnectionMaxLifetime: databaseConfig.GetDuration(constants.DatabaseConnectionMaxLifetimeInSecondsKey) * time.Second,
		ConnectionMaxIdleTime: databaseConfig.GetDuration(constants.DatabaseConnectionMaxIdleTimeInSecondsKey) * time.Second,
	})
	if err != nil {
		log.Fatal(nil).Err(err).Msg("unable to initialize database")
	}
}

func closeDatabase() {
	err := database.Close()
	if err != nil {
		log.Fatal(nil).Err(err).Msg("error closing database")
	}
}

func startRouter() {
	// get router
	router := api.GetRouter(middlewares.Logger(middlewares.LoggerMiddlewareOptions{}))
	// now start router
	err := router.Run(fmt.Sprintf(":%d", flags.Port()))
	if err != nil {
		log.Fatal(nil).Err(err).Msg("error starting router")
	}
}
