package models

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProjectsFromContext(c *gin.Context, orgIdKey string) ([]Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getProjectsFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var projects []Project

	err := DB.Preload("Organisation").Preload("Namespace").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&projects).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	fmt.Printf("getProjectsFromContext, number of projects:%d\n", len(projects))
	return projects, true
}

func GetPoliciesFromContext(c *gin.Context, orgIdKey string) ([]Policy, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getPoliciesFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var policies []Policy

	err := DB.Preload("Organisation").Preload("Namespace").Preload("Project").
		Joins("LEFT JOIN projects ON projects.id = policies.project_id").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&policies).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	fmt.Printf("getPoliciesFromContext, number of policies:%d\n", len(policies))
	return policies, true
}

func GetProjectRunsFromContext(c *gin.Context, orgIdKey string) ([]ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getProjectRunsFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var runs []ProjectRun

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).Find(&runs).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	fmt.Printf("getProjectRunsFromContext, number of runs:%d\n", len(runs))
	return runs, true
}

func GetProjectByRunId(c *gin.Context, runId uint, orgIdKey string) (*ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("GetProjectByRunId, org id: %v\n", loggedInOrganisationId)
	var projectRun ProjectRun

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("project_runs.id = ?", runId).First(&projectRun).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &projectRun, true
}

func GetProjectByProjectId(c *gin.Context, projectId uint, orgIdKey string) (*Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("GetProjectByProjectId, org id: %v\n", loggedInOrganisationId)
	var project Project

	err := DB.Preload("Organisation").Preload("Namespace").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("projects.id = ?", projectId).First(&project).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &project, true
}

func GetPolicyByPolicyId(c *gin.Context, policyId uint, orgIdKey string) (*Policy, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getPolicyByPolicyId, org id: %v\n", loggedInOrganisationId)
	var policy Policy

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Namespace").
		Joins("INNER JOIN projects ON projects.id = policies.project_id").
		Joins("INNER JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("policies.id = ?", policyId).First(&policy).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &policy, true
}

func GetDefaultNamespace(c *gin.Context, orgIdKey string) (*Namespace, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		fmt.Print("Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getDefaultNamespace, org id: %v\n", loggedInOrganisationId)
	var namespace Namespace

	err := DB.Preload("Organisation").
		Joins("INNER JOIN organisations ON namespaces.organisation_id = organisations.id").
		Where("organisations.id = ?", loggedInOrganisationId).First(&namespace).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &namespace, true
}
