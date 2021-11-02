package main

import (
	"context"
	"fmt"
	"github.com/angel-one/go-utils/log"
	"github.com/angel-one/go-utils/middlewares"
	"github.com/sinhashubham95/go-example-project/api"
	"github.com/sinhashubham95/go-example-project/constants"
	"github.com/sinhashubham95/go-example-project/utils/configs"
	"github.com/sinhashubham95/go-example-project/utils/database"
	"github.com/sinhashubham95/go-example-project/utils/flags"
	"github.com/sinhashubham95/go-example-project/utils/httpclient"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/sinhashubham95/go-example-project/docs"
)

// @title Go Example Project
// @version 1.0
// @description Go Example Project
// @termsOfService https://swagger.io/terms/

// @contact.name Shubham Sinha
// @contact.email shubham.sinha@angelbroking.com

// @BasePath /

func main() {
	ctx := context.Background()
	initConfigs(ctx)
	initLogger(ctx)
	initHTTPClient()
	initDatabase(ctx)
	defer closeDatabase(ctx)
	startRouter(ctx)
}

func initConfigs(ctx context.Context) {
	// init configs
	err := configs.Init(flags.BaseConfigPath(), constants.LoggerConfig, constants.ApplicationConfig,
		constants.DatabaseConfig)
	if err != nil {
		log.Fatal(ctx).Err(err).Msg("error initialising configs")
	}
}

func initLogger(ctx context.Context) {
	// start logger
	logLevel, err := configs.Get().GetString(constants.LoggerConfig, constants.LogLevelConfigKey)
	if err != nil {
		log.Fatal(ctx).Err(err).Msg("error getting log level")
	}
	log.InitLogger(log.Level(logLevel))
}

func initHTTPClient() {
	httpclient.Init(
		httpclient.NewRequestConfig("moxy", configs.Get().GetMapD(constants.ApplicationConfig,
			"http.moxy", nil)),
	)
}

func initDatabase(ctx context.Context) {
	// init database
	err := database.InitDatabase(ctx, database.Config{
		Server:             configs.Get().GetStringD(constants.DatabaseConfig, constants.DatabaseServerConfigKey, ""),
		Port:               int(configs.Get().GetIntD(constants.DatabaseConfig, constants.DatabasePortConfigKey, 0)),
		Name:               configs.Get().GetStringD(constants.DatabaseConfig, constants.DatabaseNameConfigKey, ""),
		Username:           configs.Get().GetStringD(constants.DatabaseConfig, constants.DatabaseUsernameConfigKey, ""),
		Password:           configs.Get().GetStringD(constants.DatabaseConfig, constants.DatabasePasswordConfigKey, ""),
		MaxOpenConnections: int(configs.Get().GetIntD(constants.DatabaseConfig, constants.DatabaseMaxOpenConnectionsKey, 0)),
		MaxIdleConnections: int(configs.Get().GetIntD(constants.DatabaseConfig, constants.DatabaseMaxIdleConnectionsKey, 0)),
		ConnectionMaxLifetime: time.Duration(configs.Get().GetIntD(constants.DatabaseConfig,
			constants.DatabaseConnectionMaxLifetimeInSecondsKey, 0)) * time.Second,
		ConnectionMaxIdleTime: time.Duration(configs.Get().GetIntD(constants.DatabaseConfig,
			constants.DatabaseConnectionMaxIdleTimeInSecondsKey, 0)) * time.Second,
	})
	if err != nil {
		log.Fatal(ctx).Err(err).Msg("unable to initialize database")
	}
}

func closeDatabase(ctx context.Context) {
	err := database.Close()
	if err != nil {
		log.Fatal(ctx).Err(err).Msg("error closing database")
	}
}

func startRouter(ctx context.Context) {
	// get router
	router := api.GetRouter(middlewares.Logger(middlewares.LoggerMiddlewareOptions{}))
	// now start router
	err := router.Run(fmt.Sprintf(":%d", flags.Port()))
	if err != nil {
		log.Fatal(ctx).Err(err).Msg("error starting router")
	}
}
