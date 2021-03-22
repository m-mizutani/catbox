package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/golambda"
	"github.com/pkg/errors"
)

var logger = golambda.Logger

const (
	contextConfig        = "config"
	contextControllerKey = "controller"
	contextRequestIDKey  = "requestID"
)

type errorResponse struct {
	Error string `json:"error"`
}

func SetupBase(engine *gin.Engine, ctrl *controller.Controller, config *Config) {
	engine.Use(func(c *gin.Context) {
		reqID := uuid.New().String()
		logger.
			With("path", c.FullPath()).
			With("params", c.Params).
			With("request_id", reqID).
			With("remote", c.ClientIP()).
			With("ua", c.Request.UserAgent()).
			Info("API request")

		c.Set(contextRequestIDKey, reqID)
		c.Set(contextControllerKey, ctrl)
		c.Set(contextConfig, config)
		c.Next()
	})

	engine.Use(func(c *gin.Context) {
		c.Next()

		if ginError := c.Errors.Last(); ginError != nil {
			if err := errors.Cause(ginError); err != nil {
				c.JSON(500, &errorResponse{Error: err.Error()})
			} else {
				c.JSON(500, &errorResponse{Error: "gin error: " + ginError.Error()})
			}
		}
	})
}

func getConfig(c *gin.Context) *Config {
	v, ok := c.Get(contextConfig)
	if !ok {
		panic("No config in contextConfig")
	}
	config, ok := v.(*Config)
	if !ok {
		panic("Type mismatch for contextConfig")
	}
	return config
}
