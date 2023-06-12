package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/alextanhongpin/go-gin-starter/usersvc"
	"github.com/gin-gonic/gin"
	"net/http"
)

func newRouter() *gin.Engine {
	r := gin.Default()
	models.ConnectDatabase()

	r.GET("/tests", controllers.FindTest)    // new
	r.POST("/tests", controllers.CreateTest) // new

	//r.Use(middleware.Cors())
	//r.Use(middleware.RequestID())

	// Setup middlewares, logger etc
	// r.Use(logger)
	// r.Use(secure)

	return r
}

func main() {
	// Setup dependencies
	cfg := config.New()
	cfg.AutomaticEnv()
	// db := database.New()
	r := gin.Default()

	authorized := r.Group("/")
	authorized.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"version":     cfg.GetString("version"),
			"deployed_at": cfg.GetString("deployed_at"),
		})
	})

	// Setup services
	usvc := usersvc.New()

	// Setup controllers
	uctl := usersvc.NewController(usvc)
	uctl.Setup(r, cfg.GetBool("usersvc_on"))

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
