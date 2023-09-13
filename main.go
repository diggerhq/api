package main

import (
	"digger.dev/cloud/controllers"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"digger.dev/cloud/services"
	"fmt"
	"github.com/alextanhongpin/go-gin-starter/config"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// based on https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
var Version = "dev"

func main() {

	log.SetOutput(os.Stdout)

	cfg := config.New()
	cfg.AutomaticEnv()
	web := controllers.WebController{Config: cfg}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:           os.Getenv("SENTRY_DSN"),
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 0.1,
		Release:          "api@" + Version,
		Debug:            true,
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v", err)
	}

	//database migrations
	models.ConnectDatabase()

	r := gin.Default()
	r.Use(sentrygin.New(sentrygin.Options{Repanic: true}))

	r.Static("/static", "./templates/static")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"build_date":  cfg.GetString("build_date"),
			"deployed_at": cfg.GetString("deployed_at"),
			"version":     Version,
			"commit_sha":  Version,
		})
	})

	r.SetFuncMap(template.FuncMap{
		"formatAsDate": func(msec int64) time.Time {
			return time.UnixMilli(msec)
		},
	})

	r.LoadHTMLGlob("templates/*.tmpl")
	r.GET("/", web.RedirectToLoginSubdomain)

	auth := services.Auth{
		HttpClient: http.Client{},
		Host:       os.Getenv("AUTH_HOST"),
		Secret:     os.Getenv("AUTH_SECRET"),
		ClientId:   os.Getenv("FRONTEGG_CLIENT_ID"),
	}

	r.POST("/github-app-webhook", controllers.GithubAppWebHook)

	githubGroup := r.Group("/github")
	githubGroup.Use(middleware.WebAuth(auth))
	githubGroup.GET("/callback", controllers.GithubAppCallbackPage)
	githubGroup.GET("/test/job", controllers.GihHubCreateTestJobPage)

	projectsGroup := r.Group("/projects")
	projectsGroup.Use(middleware.WebAuth(auth))
	projectsGroup.GET("/", web.ProjectsPage)
	projectsGroup.GET("/add", web.AddProjectPage)
	projectsGroup.POST("/add", web.AddProjectPage)
	projectsGroup.GET("/:projectid/details", web.ProjectDetailsPage)
	projectsGroup.POST("/:projectid/details", web.ProjectDetailsUpdatePage)

	runsGroup := r.Group("/runs")
	runsGroup.Use(middleware.WebAuth(auth))
	runsGroup.GET("/", web.RunsPage)
	runsGroup.GET("/:runid/details", web.RunDetailsPage)

	reposGroup := r.Group("/repo")
	reposGroup.Use(middleware.WebAuth(auth))
	reposGroup.GET("/:repoid/", web.UpdateRepoPage)
	reposGroup.POST("/:repoid/", web.UpdateRepoPage)

	policiesGroup := r.Group("/policies")
	policiesGroup.Use(middleware.WebAuth(auth))
	policiesGroup.GET("/", web.PoliciesPage)
	policiesGroup.GET("/add", web.AddPolicyPage)
	policiesGroup.POST("/add", web.AddPolicyPage)
	policiesGroup.GET("/:policyid/details", web.PolicyDetailsPage)
	policiesGroup.POST("/:policyid/details", web.PolicyDetailsUpdatePage)

	authorized := r.Group("/")
	authorized.Use(middleware.BearerTokenAuth(auth), middleware.AccessLevel(models.AccessPolicyType, models.AdminPolicyType))

	admin := r.Group("/")
	admin.Use(middleware.BearerTokenAuth(auth), middleware.AccessLevel(models.AdminPolicyType))

	fronteggWebhookProcessor := r.Group("/")
	fronteggWebhookProcessor.Use(middleware.SecretCodeAuth())

	authorized.GET("/repos/:repo/projects/:projectName/access-policy", controllers.FindAccessPolicy)
	authorized.GET("/orgs/:organisation/access-policy", controllers.FindAccessPolicyForOrg)

	authorized.GET("/repos/:repo/projects/:projectName/plan-policy", controllers.FindPlanPolicy)
	authorized.GET("/orgs/:organisation/plan-policy", controllers.FindPlanPolicyForOrg)

	authorized.GET("/repos/:repo/projects/:projectName/drift-policy", controllers.FindDriftPolicy)
	authorized.GET("/orgs/:organisation/drift-policy", controllers.FindDriftPolicyForOrg)

	authorized.GET("/repos/:repo/projects/:projectName/runs", controllers.RunHistoryForProject)
	authorized.POST("/repos/:repo/projects/:projectName/runs", controllers.CreateRunForProject)
	authorized.GET("/repos/:repo/projects", controllers.FindProjectsForRepo)
	authorized.POST("/repos/:repo/report-projects", controllers.ReportProjectsForRepo)

	authorized.GET("/orgs/:organisation/projects", controllers.FindProjectsForOrg)

	admin.PUT("/repos/:repo/projects/:projectName/access-policy", controllers.UpsertAccessPolicyForRepoAndProject)
	admin.PUT("/orgs/:organisation/access-policy", controllers.UpsertAccessPolicyForOrg)

	admin.PUT("/repos/:repo/projects/:projectName/plan-policy", controllers.UpsertPlanPolicyForRepoAndProject)
	admin.PUT("/orgs/:organisation/plan-policy", controllers.UpsertPlanPolicyForOrg)

	admin.PUT("/repos/:repo/projects/:projectName/drift-policy", controllers.UpsertDriftPolicyForRepoAndProject)
	admin.PUT("/orgs/:organisation/drift-policy", controllers.UpsertDriftPolicyForOrg)

	admin.POST("/tokens/issue-access-token", controllers.IssueAccessTokenForOrg)

	fronteggWebhookProcessor.POST("/create-org-from-frontegg", controllers.CreateFronteggOrgFromWebhook)

	r.Run(fmt.Sprintf(":%d", cfg.GetInt("port")))
}
