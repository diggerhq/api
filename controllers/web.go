package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WebController struct {
	Config *config.Config
}

func (web *WebController) MainPage(c *gin.Context) {
	url := web.Config.Get("FRONTEGG_URL")
	clientId := web.Config.Get("FRONTEGG_CLIENT_ID")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"FronteggUrl":      url,
		"FronteggClientId": clientId,
	})
}

func (web *WebController) ProjectsPage(c *gin.Context) {
	projects, done := web.getProjects(c)
	if done {
		return
	}

	//org := &models.Organisation{Name: "digger"}
	//namespace := &models.Namespace{Name: "main"}

	//projects := make([]models.Project, 0)
	//projects = append(projects, models.Project{Name: "aaaa", Organisation: org, Namespace: namespace})
	//projects = append(projects, models.Project{Name: "bbbb", Organisation: org, Namespace: namespace})

	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}

func (web *WebController) getProjects(c *gin.Context) ([]models.Project, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	fmt.Printf("getProjects, org id %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, true
	}

	var projects []models.Project

	err := models.DB.Preload("Organisation").Preload("Namespace").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&projects).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, true
	}
	return projects, false
}

func (web *WebController) RunsPage(c *gin.Context) {
	org := &models.Organisation{Name: "digger"}
	namespace := &models.Namespace{Name: "main"}
	project := &models.Project{Name: "Test Project", Organisation: org, Namespace: namespace}
	runs := make([]models.ProjectRun, 0)
	runs = append(runs, models.ProjectRun{Project: project, Status: "ok", Output: "test output", Command: "echo"})
	runs = append(runs, models.ProjectRun{Project: project, Status: "failed", Output: "test output", Command: "ls"})

	c.HTML(http.StatusOK, "runs.tmpl", gin.H{
		"Runs": runs,
	})
}

func (web *WebController) PoliciesPage(c *gin.Context) {
	org := &models.Organisation{Name: "digger"}
	namespace := &models.Namespace{Name: "main"}
	project := &models.Project{Name: "Test Project", Organisation: org, Namespace: namespace}

	policies := make([]models.Policy, 0)
	policies = append(policies, models.Policy{Project: project, Organisation: org, Namespace: namespace})
	policies = append(policies, models.Policy{Project: project, Organisation: org, Namespace: namespace})

	fmt.Println("policies.tmpl")
	c.HTML(http.StatusOK, "policies.tmpl", gin.H{
		"Policies": policies,
	})
}

func (web *WebController) PolicyDetailsPage(c *gin.Context) {
	org := &models.Organisation{Name: "digger"}
	namespace := &models.Namespace{Name: "main"}
	project := &models.Project{Name: "Test Project", Organisation: org, Namespace: namespace}

	policy := models.Policy{Project: project, Organisation: org, Namespace: namespace}

	fmt.Println("policy_details.tmpl")
	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy": policy,
	})
}

func (web *WebController) ProjectDetailsPage(c *gin.Context) {
	org := &models.Organisation{Name: "digger"}
	namespace := &models.Namespace{Name: "main"}
	project := &models.Project{Name: "Test Project", Organisation: org, Namespace: namespace}
	fmt.Println("project_details.tmpl")
	c.HTML(http.StatusOK, "project_details.tmpl", gin.H{
		"Project": project,
	})
}
