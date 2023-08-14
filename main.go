package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"digger.dev/cloud/services"
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

	web := controllers.WebController{Config: cfg}

	//database migrations
	models.ConnectDatabase()

	r := gin.Default()

	r.Static("/static", "./templates/static")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"deployed_at": cfg.GetString("deployed_at"),
			"version":     Version,
			"commit_sha":  os.Getenv("COMMIT_SHA"),
		})
	})

	r.LoadHTMLFiles("templates/index.tmpl", "templates/projects.tmpl")
	r.GET("/", web.MainPage)
	r.GET("/oauth/callback", web.MainPage)
	auth := services.Auth{
		HttpClient: http.Client{},
		Host:       os.Getenv("AUTH_HOST"),
		Secret:     os.Getenv("AUTH_SECRET"),
		ClientId:   os.Getenv("FRONTEGG_CLIENT_ID"),
	}
	webGroup := r.Group("/projects")
	webGroup.Use(middleware.WebAuth(auth))
	webGroup.GET("/", web.ProjectsPage)

	authorized := r.Group("/")
	authorized.Use(middleware.BearerTokenAuth(auth), middleware.AccessLevel(models.AccessPolicyType, models.AdminPolicyType))

	admin := r.Group("/")
	admin.Use(middleware.BearerTokenAuth(auth), middleware.AccessLevel(models.AdminPolicyType))

	fronteggWebhookProcessor := r.Group("/")
	fronteggWebhookProcessor.Use(middleware.SecretCodeAuth())

	authorized.GET("/repos/:namespace/projects/:projectName/access-policy", controllers.FindAccessPolicy)
	authorized.GET("/orgs/:organisation/access-policy", controllers.FindAccessPolicyForOrg)

	authorized.GET("/repos/:namespace/projects/:projectName/plan-policy", controllers.FindPlanPolicy)
	authorized.GET("/orgs/:organisation/plan-policy", controllers.FindPlanPolicyForOrg)

	authorized.GET("/repos/:namespace/projects/:projectName/drift-policy", controllers.FindDriftPolicy)
	authorized.GET("/orgs/:organisation/drift-policy", controllers.FindDriftPolicyForOrg)

	authorized.GET("/repos/:namespace/projects/:projectName/runs", controllers.RunHistoryForProject)
	authorized.POST("/repos/:namespace/projects/:projectName/runs", controllers.CreateRunForProject)
	authorized.GET("/repos/:namespace/projects", controllers.FindProjectsForNamespace)
	authorized.POST("/repos/:namespace/report-projects", controllers.ReportProjectsForNamespace)

	authorized.GET("/orgs/:organisation/projects", controllers.FindProjectsForOrg)

	admin.PUT("/repos/:namespace/projects/:projectName/access-policy", controllers.UpsertAccessPolicyForNamespaceAndProject)
	admin.PUT("/orgs/:organisation/access-policy", controllers.UpsertAccessPolicyForOrg)

	admin.PUT("/repos/:namespace/projects/:projectName/plan-policy", controllers.UpsertPlanPolicyForNamespaceAndProject)
	admin.PUT("/orgs/:organisation/plan-policy", controllers.UpsertPlanPolicyForOrg)

	admin.PUT("/repos/:namespace/projects/:projectName/drift-policy", controllers.UpsertDriftPolicyForNamespaceAndProject)
	admin.PUT("/orgs/:organisation/drift-policy", controllers.UpsertDriftPolicyForOrg)

	admin.POST("/tokens/issue-access-token", controllers.IssueAccessTokenForOrg)

	fronteggWebhookProcessor.POST("/create-org-from-frontegg", controllers.CreateFronteggOrgFromWebhook)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
