package api

import (
	"github.com/gin-gonic/gin"
)

const (
	contextControllerKey = "controller"
	contextRequestIDKey  = "requestID"
)

// SetupAPI sets API handlers to gin.RouterGroup
func SetupAPI(engine *gin.Engine) {
	r := engine.Group("/api/v1")

	// Repository
	r.GET("/repository", getRepositories)
	r.GET("/repository/:registry/:repo", getRepository)
	r.GET("/repository/:registry/:repo/config", getRepositoryConfig)
	r.PUT("/repository/:registry/:repo/config", updateRepositoryConfig)

	// Team
	r.GET("/team", getTeams)
	r.GET("/team/:team_id", getTeam)
	r.GET("/team/:team_id/repository", getTeamRepository)
	r.POST("/team", createTeam)
	r.PUT("/team/:team", updateTeam)

	// Vulnerability
	r.GET("/vuln", getVulnerabilities)
	r.GET("/vuln/:vuln_id", getVulnerability)
}

// Repository
func getRepositories(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func getRepository(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func getRepositoryConfig(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func updateRepositoryConfig(c *gin.Context) {
	c.JSON(200, struct{}{})
}

// Team
func getTeams(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func getTeam(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func getTeamRepository(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func createTeam(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func updateTeam(c *gin.Context) {
	c.JSON(200, struct{}{})
}

// Vulnerability
func getVulnerabilities(c *gin.Context) {
	c.JSON(200, struct{}{})
}
func getVulnerability(c *gin.Context) {
	c.JSON(200, struct{}{})
}
