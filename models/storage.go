package models

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func GetProjectsFromContext(c *gin.Context, orgIdKey string) ([]Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getProjectsFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var projects []Project

	err := DB.Preload("Organisation").Preload("Repo").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
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

	err := DB.Preload("Organisation").Preload("Repo").Preload("Project").
		Joins("LEFT JOIN projects ON projects.id = policies.project_id").
		Joins("LEFT JOIN repos ON projects.repo_id = repos.id").
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

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
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

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
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

	err := DB.Preload("Organisation").Preload("Repo").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
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

	err := DB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
		Joins("INNER JOIN projects ON projects.id = policies.project_id").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("policies.id = ?", policyId).First(&policy).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &policy, true
}

func GetDefaultRepo(c *gin.Context, orgIdKey string) (*Repo, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		fmt.Print("Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getDefaultRepo, org id: %v\n", loggedInOrganisationId)
	var repo Repo

	err := DB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ?", loggedInOrganisationId).First(&repo).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &repo, true
}

func GetRepo(c *gin.Context, orgIdKey string, repoId uint) (*Repo, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		fmt.Print("Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getDefaultRepo, org id: %v\n", loggedInOrganisationId)
	var repo Repo

	err := DB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ? AND repos.id=?", loggedInOrganisationId, repoId).First(&repo).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &repo, true
}

func GitHubRepoAdded(installationId int64, appId int, login string, accountId int64, repoFullName string) error {
	// check if item exist already
	item := GithubAppInstallation{}
	result := DB.Where("github_installation_id = ? AND repo=?", installationId, repoFullName).First(&item)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to find github installation in database. %v", result.Error)
		}
	}

	if result.RowsAffected == 0 {
		item := GithubAppInstallation{
			GithubInstallationId: installationId,
			GithubAppId:          int64(appId),
			Login:                login,
			AccountId:            int(accountId),
			Repo:                 repoFullName,
			State:                Active,
		}
		err := DB.Create(&item).Error
		if err != nil {
			fmt.Printf("Failed to save github installation item to database. %v\n", err)
			return fmt.Errorf("failed to save github installation item to database. %v", err)
		}
	} else {
		fmt.Printf("Record for installation_id: %d, repo: %s, with state=active exist already.", installationId, repoFullName)
		item.State = Active
		item.UpdatedAt = time.Now()
		err := DB.Save(item).Error
		if err != nil {
			return fmt.Errorf("failed to update github installation in the database. %v", err)
		}
	}
	return nil
}

func GitHubRepoRemoved(installationId int64, appId int, repoFullName string) error {
	item := GithubAppInstallation{}
	err := DB.Where("github_installation_id = ? AND state=? AND github_app_id=? AND repo=?", installationId, Active, appId, repoFullName).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Record not found for installationId: %d, state=active, githubAppId: %d and repo: %s", installationId, appId, repoFullName)
			return nil
		}
		return fmt.Errorf("failed to find github installation in database. %v", err)
	}
	item.State = Deleted
	item.UpdatedAt = time.Now()
	err = DB.Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update github installation in the database. %v", err)
	}
	return nil
}
