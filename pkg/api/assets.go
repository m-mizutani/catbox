package api

import (
	"github.com/gin-gonic/gin"
)

// SetupAssets sets routes to provide static assets
func SetupAssets(engine *gin.Engine) {
	engine.GET("/", getIndex)
	engine.GET("/js/bundle.js", getBundleJS)
}

// Assets
func getIndex(c *gin.Context) {
	c.Data(200, "text/html", []byte(`<html><body>meow!</body></html>`))
}
func getBundleJS(c *gin.Context) {
	c.Data(200, "application/javascript", []byte(`console.log('meow');`))
}
