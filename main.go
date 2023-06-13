package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/gin-gonic/gin"
	"net/http"
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
	authorized := r.Group("/")
	authorized.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"version":     cfg.GetString("version"),
			"deployed_at": cfg.GetString("deployed_at"),
		})
	})

	r.GET("/tests", controllers.FindTest)
	r.POST("/tests", controllers.CreateTest)

	r.GET("/policies", controllers.FindPolicy)
	r.POST("/policies", controllers.CreatePolicy)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
