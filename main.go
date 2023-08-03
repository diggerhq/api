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

	r.LoadHTMLGlob("templates/*.tmpl")
	r.GET("/", web.MainPage)
	r.GET("/oauth/callback", web.MainPage)

	projectsGroup := r.Group("/projects")
	projectsGroup.Use(middleware.WebAuth())
	projectsGroup.GET("/", web.ProjectsPage)
	projectsGroup.GET("/:projectid/details", web.ProjectDetailsPage)
	projectsGroup.POST("/:projectid/details", web.ProjectDetailsUpdatePage)

	runsGroup := r.Group("/runs")
	runsGroup.Use(middleware.WebAuth())
	runsGroup.GET("/", web.RunsPage)
	runsGroup.GET("/:runid/details", web.RunDetailsPage)

	policiesGroup := r.Group("/policies")
	policiesGroup.Use(middleware.WebAuth())
	policiesGroup.GET("/", web.PoliciesPage)
	policiesGroup.GET("/:policyid/details", web.PolicyDetailsPage)
	//policiesGroup.POST("/:policyid/details", web.PolicyDetailsUpdatePage)

	authorized := r.Group("/")
	authorized.Use(middleware.BearerTokenAuth(), middleware.AccessLevel(models.AccessPolicyType, models.AdminPolicyType))

	admin := r.Group("/")
	admin.Use(middleware.BearerTokenAuth(), middleware.AccessLevel(models.AdminPolicyType))

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

	authorized.GET("/orgs/:organisation/projects", controllers.FindProjectsForOrg)
	authorized.POST("/orgs/:organisation/report-projects", controllers.ReportProjectsForOrg)

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
