package main

import (
	cloud_config "digger.dev/cloud/config"
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

// based on https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
var Version = "dev"

func main() {
	cfg := config.New()
	cfg.AutomaticEnv()

	var envVars cloud_config.EnvVariables

	if err := env.Parse(&envVars); err != nil {
		fmt.Printf("%+v\n", err)
	}

	//database migrations
	models.ConnectDatabase(&envVars)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"deployed_at": cfg.GetString("deployed_at"),
			"version":     Version,
			"commit_sha":  os.Getenv("COMMIT_SHA"),
		})
	})

	r.POST("/github-app-callback", controllers.GitHubAppCallback())

	r.POST("/github-app-webhook", controllers.GitHubAppWebHook())

	authorized := r.Group("/")
	authorized.Use(middleware.BearerTokenAuth(&envVars), middleware.AccessLevel(models.AccessPolicyType, models.AdminPolicyType))

	admin := r.Group("/")
	admin.Use(middleware.BearerTokenAuth(&envVars), middleware.AccessLevel(models.AdminPolicyType))

	fronteggWebhookProcessor := r.Group("/")
	fronteggWebhookProcessor.Use(middleware.SecretCodeAuth(&envVars))

	authorized.GET("/repos/:namespace/projects/:projectName/access-policy", controllers.FindPolicy)
	authorized.GET("/orgs/:organisation/access-policy", controllers.FindPolicyForOrg)

	admin.PUT("/repos/:namespace/projects/:projectName/access-policy", controllers.UpsertPolicyForNamespaceAndProject)
	admin.PUT("/orgs/:organisation/access-policy", controllers.UpsertPolicyForOrg)
	admin.POST("/tokens/issue-access-token", controllers.IssueAccessTokenForOrg)

	fronteggWebhookProcessor.POST("/create-org-from-frontegg", controllers.CreateFronteggOrgFromWebhook)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
