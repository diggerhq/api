package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// based on https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
var Version = "dev"

func main() {
	cfg := config.New()
	cfg.AutomaticEnv()

	//database migrations
	models.ConnectDatabase()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"deployed_at": cfg.GetString("deployed_at"),
			"version":     Version,
			"commit_sha":  os.Getenv("COMMIT_SHA"),
		})
	})

	authorized := r.Group("/")
	authorized.Use(middleware.BasicBearerTokenAuth(), middleware.AccessLevel(models.AccessPolicyType, models.AdminPolicyType))

	admin := r.Group("/")
	admin.Use(middleware.BasicBearerTokenAuth(), middleware.AccessLevel(models.AdminPolicyType))

	authorized.GET("/repos/:namespace/projects/:projectName/access-policy", controllers.FindPolicy)
	authorized.GET("/orgs/:organisation/access-policy", controllers.FindPolicyForOrg)

	admin.PUT("/repos/:namespace/projects/:projectName/access-policy", controllers.UpsertPolicyForNamespaceAndProject)
	admin.PUT("/orgs/:organisation/access-policy", controllers.UpsertPolicyForOrg)
	admin.POST("/tokens/issue-access-token", controllers.IssueAccessTokenForOrg)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
