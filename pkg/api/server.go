package api

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func SetupBase(engine *gin.Engine, ctrl *controller.Controller) {
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
		c.Next()
	})
}
