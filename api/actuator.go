package api

import (
	"github.com/sinhashubham95/go-example-project/constants"
	"github.com/sinhashubham95/go-example-project/utils/flags"
	"github.com/gin-gonic/gin"
	goActuator "github.com/sinhashubham95/go-actuator"
)

var (
	actuatorHandler = goActuator.GetActuatorHandler(&goActuator.Config{
		Env:     flags.Env(),
		Name:    constants.ApplicationName,
		Port:    flags.Port(),
		Version: "",
	})
)

func actuator(ctx *gin.Context) {
	actuatorHandler(ctx.Writer, ctx.Request)
}
