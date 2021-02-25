package main

import (
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

const (
	contextRequestIDKey = "requestID"
)

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		var req events.APIGatewayProxyRequest
		if err := event.Bind(&req); err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()

		r.Use(func(c *gin.Context) {
			reqID := uuid.New().String()
			logger.
				With("path", c.FullPath()).
				With("params", c.Params).
				With("request_id", reqID).
				With("remote", c.ClientIP()).
				With("ua", c.Request.UserAgent()).
				Info("API request")

			c.Set(contextRequestIDKey, reqID)
			c.Next()
		})

		v1 := r.Group("/api/v1")
		v1.GET("/repository", func(c *gin.Context) {
			// TODO:
		})

		ginLambda := ginadapter.New(r)

		return ginLambda.Proxy(req)
	})
}
