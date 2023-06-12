package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/models"
	"digger.dev/cloud/platform/authenticator"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/gin-gonic/gin"
	"log"
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
	models.ConnectDatabase()

	cfg := config.New()
	cfg.AutomaticEnv()
	// db := database.New()
	r := gin.Default()

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	authorized := r.Group("/")
	authorized.Use(auth.AuthRequired())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"version":     cfg.GetString("version"),
			"deployed_at": cfg.GetString("deployed_at"),
		})
	})

	authorized.GET("/tests", controllers.FindTest)
	authorized.POST("/tests", controllers.CreateTest)

	r.GET("/callback", controllers.AuthHandler())

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
