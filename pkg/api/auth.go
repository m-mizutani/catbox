package api

import (
	"github.com/gin-gonic/gin"
)

// SetupAuth sets routes for authentication
func SetupAuth(engine *gin.Engine) {
	// Auth
	r := engine.Group("/auth")
	r.GET("/", getAuth)
	r.GET("/google", authGoogle)
	r.GET("/google/callback", authGoogleCallback)
	r.GET("/logout", authLogout)
}

// Auth
func getAuth(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func authGoogle(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func authGoogleCallback(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func authLogout(c *gin.Context) {
	c.JSON(200, struct{}{})
}
