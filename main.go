package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func newRouter() *gin.Engine {
	r := gin.Default()
	models.ConnectDatabase()

	//r.Use(middleware.Cors())
	//r.Use(middleware.RequestID())

	// Setup middlewares, logger etc
	// r.Use(logger)
	// r.Use(secure)

	return r
}

func main() {
	cfg := config.New()
	cfg.AutomaticEnv()

	//database migrations
	models.ConnectDatabase()

	r := gin.Default()

	version, _ := os.ReadFile("version.txt")
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"deployed_at": cfg.GetString("deployed_at"),
			"version":     string(version),
		})
	})

	authorized := r.Group("/")
	authorized.GET("/tests", controllers.FindTest)
	authorized.POST("/tests", controllers.CreateTest)

	authorized.GET("/repos/:namespace/projects/:projectName/access-policy", controllers.FindPolicy)
	authorized.PUT("/repos/:namespace/projects/:projectName/access-policy", controllers.UpdatePolicy)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
