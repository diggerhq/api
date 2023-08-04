package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type WebController struct {
	Config *config.Config
}

func (web *WebController) validateRequestProjectId(c *gin.Context) (*models.Project, bool) {
	projectId64, err := strconv.ParseUint(c.Param("projectid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse project id")
		return nil, false
	}
	projectId := uint(projectId64)
	projects, done := web.getProjectsFromContext(c)
	if !done {
		return nil, false
	}

	for _, p := range projects {
		if projectId == p.ID {
			return &p, true
		}
	}

	c.String(http.StatusForbidden, "Not allowed to access this resource")
	return nil, false
}

func (web *WebController) getProjectsFromContext(c *gin.Context) ([]models.Project, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	fmt.Printf("getProjectsFromContext, org id %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var projects []models.Project

	err := models.DB.Preload("Organisation").Preload("Namespace").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&projects).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, false
	}

	fmt.Printf("getProjectsFromContext, number of projects:%d\n", len(projects))
	return projects, true
}

func (web *WebController) getPoliciesFromContext(c *gin.Context) ([]models.Policy, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	fmt.Printf("getPoliciesFromContext, org id %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var policies []models.Policy

	err := models.DB.Preload("Organisation").Preload("Namespace").Preload("Project").
		Joins("LEFT JOIN projects ON projects.id = policies.project_id").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&policies).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, false
	}

	fmt.Printf("getPoliciesFromContext, number of policies:%d\n", len(policies))
	return policies, true
}

func (web *WebController) getProjectRunsFromContext(c *gin.Context) ([]models.ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	fmt.Printf("getProjectRunsFromContext, org id %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var runs []models.ProjectRun

	err := models.DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&runs).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, false
	}

	fmt.Printf("getProjectRunsFromContext, number of runs:%d\n", len(runs))
	return runs, true
}

func (web *WebController) getProjectByRunId(c *gin.Context, runId uint) (*models.ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getProjectsFromContext, org id %v\n", loggedInOrganisationId)
	var projectRun models.ProjectRun

	err := models.DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("project_runs.id = ?", runId).First(&projectRun).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, false
	}

	return &projectRun, true
}

func (web *WebController) getPolicyByPolicyId(c *gin.Context, policyId uint) (*models.Policy, bool) {
	loggedInOrganisationId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getPolicyByPolicyId, org id %v\n", loggedInOrganisationId)
	var policy models.Policy

	err := models.DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = policies.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("policies.id = ?", policyId).First(&policy).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return nil, false
	}

	return &policy, true
}

func (web *WebController) ProjectsPage(c *gin.Context) {
	projects, done := web.getProjectsFromContext(c)
	if !done {
		return
	}

	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}

func (web *WebController) RunsPage(c *gin.Context) {
	runs, done := web.getProjectRunsFromContext(c)
	if !done {
		return
	}
	c.HTML(http.StatusOK, "runs.tmpl", gin.H{
		"Runs": runs,
	})
}

func (web *WebController) PoliciesPage(c *gin.Context) {
	policies, done := web.getPoliciesFromContext(c)
	if !done {
		return
	}
	fmt.Println("policies.tmpl")
	c.HTML(http.StatusOK, "policies.tmpl", gin.H{
		"Policies": policies,
	})
}

func (web *WebController) PolicyDetailsPage(c *gin.Context) {
	policyId64, err := strconv.ParseUint(c.Param("policyid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse policy id")
		return
	}
	policyId := uint(policyId64)
	policy, ok := web.getPolicyByPolicyId(c, policyId)
	if !ok {
		return
	}

	fmt.Println("policy_details.tmpl")
	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy": policy,
	})
}

func (web *WebController) ProjectDetailsPage(c *gin.Context) {
	project, ok := web.validateRequestProjectId(c)
	if !ok {
		return
	}

	fmt.Println("project_details.tmpl")
	c.HTML(http.StatusOK, "project_details.tmpl", gin.H{
		"Project": project,
	})
}

func (web *WebController) RunDetailsPage(c *gin.Context) {
	runId64, err := strconv.ParseUint(c.Param("runid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse project run id")
		return
	}
	runId := uint(runId64)
	run, ok := web.getProjectByRunId(c, runId)
	if !ok {
		return
	}

	fmt.Println("run_details.tmpl")
	c.HTML(http.StatusOK, "run_details.tmpl", gin.H{
		"Run": run,
	})
}

func (web *WebController) ProjectDetailsUpdatePage(c *gin.Context) {
	project, ok := web.validateRequestProjectId(c)
	if !ok {
		return
	}

	message := ""
	projectName := c.PostForm("project_name")
	if projectName != project.Name {
		project.Name = projectName
		models.DB.Save(project)
		fmt.Printf("project name has been updated to %s\n", projectName)
		message = "Project has been updated successfully"
	}

	c.HTML(http.StatusOK, "project_details.tmpl", gin.H{
		"Project": project,
		"Message": message,
	})
}

func (web *WebController) PolicyDetailsUpdatePage(c *gin.Context) {
	policyId64, err := strconv.ParseUint(c.Param("policyid"), 10, 32)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to parse policy id")
		return
	}
	policyId := uint(policyId64)
	policy, ok := web.getPolicyByPolicyId(c, policyId)
	if !ok {
		return
	}

	message := ""
	policyText := c.PostForm("policy")
	fmt.Printf("policyText: %v\n", policyText)
	if policyText != policy.Policy {
		policy.Policy = policyText
		models.DB.Save(policy)
		fmt.Printf("Policy has been updated. policy id: %v", policy.ID)
		message = "Policy has been updated successfully"
	}

	c.HTML(http.StatusOK, "policy_details.tmpl", gin.H{
		"Policy":  policy,
		"Message": message,
	})
}

func (web *WebController) RedirectToLoginSubdomain(context *gin.Context) {
	host := context.Request.Host
	hostParts := strings.Split(host, ".")
	if len(hostParts) > 2 {
		hostParts[0] = "login"
		host = strings.Join(hostParts, ".")
	}
	context.Redirect(http.StatusMovedPermanently, fmt.Sprintf("https://%s", host))
}
