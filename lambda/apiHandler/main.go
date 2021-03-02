package main

import (
	"github.com/aws/aws-lambda-go/events"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"

	"github.com/m-mizutani/catbox/pkg/api"
	"github.com/m-mizutani/catbox/pkg/controller"
	"github.com/m-mizutani/golambda"
)

var logger = golambda.Logger

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		var req events.APIGatewayProxyRequest
		if err := event.Bind(&req); err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		ctrl := controller.New()

		gin.SetMode(gin.ReleaseMode)
		engine := gin.Default()
		api.SetupBase(engine, ctrl)
		api.SetupAssets(engine)
		api.SetupAPI(engine)

		ginLambda := ginadapter.New(engine)

		return ginLambda.Proxy(req)
	})
}
