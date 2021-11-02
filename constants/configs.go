package constants

// config names
const (
	LoggerConfig      = "logger"
	ApplicationConfig = "application"
	DatabaseConfig    = "database"
	CounterConfig     = "counter"
)

// config keys
const (
	LogLevelConfigKey                         = "level"
	DatabaseServerConfigKey                   = "server"
	DatabasePortConfigKey                     = "port"
	DatabaseNameConfigKey                     = "name"
	DatabaseUsernameConfigKey                 = "username"
	DatabasePasswordConfigKey                 = "password"
	DatabaseMaxOpenConnectionsKey             = "maxOpenConnections"
	DatabaseMaxIdleConnectionsKey             = "maxIdleConnections"
	DatabaseConnectionMaxLifetimeInSecondsKey = "connectionMaxLifetimeInSeconds"
	DatabaseConnectionMaxIdleTimeInSecondsKey = "connectionMaxIdleTimeInSeconds"
	CounterQueryTimeoutInMillisKey            = "queryTimeoutInMillis"
)
