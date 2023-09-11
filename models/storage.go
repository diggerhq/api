package models

import (
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"time"
)

func (db *Database) GetProjectsFromContext(c *gin.Context, orgIdKey string) ([]Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getProjectsFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var projects []Project

	err := db.GormDB.Preload("Organisation").Preload("Repo").
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

func (db *Database) GetPoliciesFromContext(c *gin.Context, orgIdKey string) ([]Policy, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getPoliciesFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var policies []Policy

	err := db.GormDB.Preload("Organisation").Preload("Repo").Preload("Project").
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

func (db *Database) GetProjectRunsFromContext(c *gin.Context, orgIdKey string) ([]ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	fmt.Printf("getProjectRunsFromContext, org id: %v\n", loggedInOrganisationId)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	var runs []ProjectRun

	err := db.GormDB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
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

func (db *Database) GetProjectByRunId(c *gin.Context, runId uint, orgIdKey string) (*ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("GetProjectByRunId, org id: %v\n", loggedInOrganisationId)
	var projectRun ProjectRun

	err := db.GormDB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
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

func (db *Database) GetProjectByProjectId(c *gin.Context, projectId uint, orgIdKey string) (*Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("GetProjectByProjectId, org id: %v\n", loggedInOrganisationId)
	var project Project

	err := db.GormDB.Preload("Organisation").Preload("Repo").
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

func (db *Database) GetPolicyByPolicyId(c *gin.Context, policyId uint, orgIdKey string) (*Policy, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getPolicyByPolicyId, org id: %v\n", loggedInOrganisationId)
	var policy Policy

	err := db.GormDB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
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

func (db *Database) GetDefaultRepo(c *gin.Context, orgIdKey string) (*Repo, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		fmt.Print("Not allowed to access this resource")
		return nil, false
	}

	fmt.Printf("getDefaultRepo, org id: %v\n", loggedInOrganisationId)
	var repo Repo

	err := db.GormDB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ?", loggedInOrganisationId).First(&repo).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &repo, true
}

// GetRepo returns digger repo by organisationId and repo name (diggerhq-digger)
func (db *Database) GetRepo(orgIdKey any, repoName string) (*Repo, bool) {

	fmt.Printf("getDefaultRepo, org id: %v\n", orgIdKey)
	var repo Repo

	err := db.GormDB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ? AND repos.name=?", orgIdKey, repoName).First(&repo).Error

	if err != nil {
		fmt.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &repo, true
}

func (db *Database) GitHubRepoAdded(installationId int64, appId int, login string, accountId int64, repoFullName string) error {
	app := GithubApp{}

	// todo: do we need to create a github app here
	result := db.GormDB.Where(&app, GithubApp{GithubId: int64(appId)}).FirstOrCreate(&app)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to create github app in database. %v", result.Error)
		}
	}

	// check if item exist already
	item := GithubAppInstallation{}
	result = db.GormDB.Where("github_installation_id = ? AND repo=? AND github_app_id=?", installationId, repoFullName, appId).First(&item)
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
		err := db.GormDB.Create(&item).Error
		if err != nil {
			fmt.Printf("Failed to save github installation item to database. %v\n", err)
			return fmt.Errorf("failed to save github installation item to database. %v", err)
		}
	} else {
		fmt.Printf("Record for installation_id: %d, repo: %s, with state=active exist already.", installationId, repoFullName)
		item.State = Active
		item.UpdatedAt = time.Now()
		err := db.GormDB.Save(item).Error
		if err != nil {
			return fmt.Errorf("failed to update github installation in the database. %v", err)
		}
	}
	return nil
}

func (db *Database) GitHubRepoRemoved(installationId int64, appId int, repoFullName string) error {
	item := GithubAppInstallation{}
	err := db.GormDB.Where("github_installation_id = ? AND state=? AND github_app_id=? AND repo=?", installationId, Active, appId, repoFullName).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Record not found for installationId: %d, state=active, githubAppId: %d and repo: %s", installationId, appId, repoFullName)
			return nil
		}
		return fmt.Errorf("failed to find github installation in database. %v", err)
	}
	item.State = Deleted
	item.UpdatedAt = time.Now()
	err = db.GormDB.Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update github installation in the database. %v", err)
	}
	return nil
}

func (db *Database) GetGitHubAppInstallationByOrgAndRepo(orgId any, repo string) (*GithubAppInstallation, error) {
	link, err := db.GetGitHubInstallationLinkForOrg(orgId)
	if err != nil {
		return nil, err
	}

	installation := GithubAppInstallation{}
	result := db.GormDB.Where("github_installation_id = ? AND state=? AND repo=?", link.GithubInstallationId, GithubAppInstallationLinkActive, repo).Find(&installation)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if installation.Model.ID == 0 {
		return nil, nil
	}
	return &installation, nil
}

// GetGitHubAppInstallationByIdAndRepo repoFullName should be in the following format: org/repo_name, for example "diggerhq/github-job-scheduler"
func (db *Database) GetGitHubAppInstallationByIdAndRepo(installationId int64, repoFullName string) (*GithubAppInstallation, error) {
	installation := GithubAppInstallation{}
	result := db.GormDB.Where("github_installation_id = ? AND state=? AND repo=?", installationId, GithubAppInstallationLinkActive, repoFullName).Find(&installation)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if installation.Model.ID == 0 {
		return nil, nil
	}
	return &installation, nil
}

// GetGitHubAppInstallationLinkByIdAndRepo repoFullName should be in the following format: org/repo_name, for example "diggerhq/github-job-scheduler"
func (db *Database) GetGitHubAppInstallationLinkByIdAndRepo(installationId int64, repoFullName string) (*GithubAppInstallationLink, error) {
	var link *GithubAppInstallationLink
	result := db.GormDB.Where("github_installation_id = ? AND state=? AND repo=?", installationId, GithubAppInstallationLinkActive, repoFullName).Find(link)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if link.Model.ID == 0 {
		return nil, nil
	}
	return link, nil
}

// GetGitHubApp
func (db *Database) GetGitHubApp(gitHubAppId int64) (*GithubApp, error) {
	app := GithubApp{}
	result := db.GormDB.Where("github_id = ?").Find(&app)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if app.Model.ID == 0 {
		return nil, nil
	}
	return &app, nil
}

func (db *Database) CreateGitHubInstallationLink(orgId uint, installationId int64) (*GithubAppInstallationLink, error) {
	l := GithubAppInstallationLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("github_installation_id = ? AND status=?", installationId, GithubAppInstallationLinkActive).Find(&l)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	if result.RowsAffected > 0 {
		if l.OrganisationId != orgId {
			return nil, fmt.Errorf("GitHub app installation %v already linked to another org ", installationId)
		}
		// record already exist, do nothing
		return &l, nil
	}

	list := []GithubAppInstallationLink{}
	// if there are other installation for this org, we need to make them inactive
	result = db.GormDB.Where("github_installation_id <> ? AND organisation_id = ? AND status=?", installationId, orgId, GithubAppInstallationLinkActive).Find(&list)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	for _, item := range list {
		item.Status = GithubAppInstallationLinkInactive
		db.GormDB.Save(item)
	}

	link := GithubAppInstallationLink{OrganisationId: orgId, GithubInstallationId: installationId, Status: GithubAppInstallationLinkActive}
	result = db.GormDB.Save(&link)
	if result.Error != nil {
		return nil, result.Error
	}
	return &link, nil
}

func (db *Database) GetGitHubInstallationLinkForOrg(orgId any) (*GithubAppInstallationLink, error) {
	l := GithubAppInstallationLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("organisation_id = ? AND status=?", orgId, GithubAppInstallationLinkActive).Find(&l)
	if result.Error != nil {
		return nil, result.Error
	}
	return &l, nil
}

func (db *Database) CreateDiggerJobLink(repoFullName string) (*GithubDiggerJobLink, error) {
	jobLink := GithubDiggerJobLink{}
	diggerJobId := uniuri.New()
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("digger_job_id = ? AND repo_full_name=? ", diggerJobId, repoFullName).Find(&jobLink)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	if result.RowsAffected > 0 {
		//if jobLink.GithubJobId != org.ID {
		//	return nil, fmt.Errorf("GitHub app installation %v already linked to another org ", installation.ID)
		//}
		// record already exist, do nothing
		return &jobLink, nil
	}

	link := GithubDiggerJobLink{Status: DiggerJobCreated, DiggerJobId: diggerJobId, RepoFullName: repoFullName}
	result = db.GormDB.Save(&link)
	if result.Error != nil {
		return nil, result.Error
	}
	return &link, nil
}

func (db *Database) UpdateDiggerJobLink(repoFullName string, diggerJobId string, githubJobId int64) (*GithubDiggerJobLink, error) {
	jobLink := GithubDiggerJobLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("digger_job_id = ? AND repo_full_name=? ", diggerJobId, repoFullName).Find(&jobLink)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	if result.RowsAffected == 1 {
		jobLink.GithubJobId = githubJobId
		result = db.GormDB.Save(&jobLink)
		if result.Error != nil {
			return nil, result.Error
		}
		return &jobLink, nil
	}
	return &jobLink, nil
}

func (db *Database) GetOrganisationById(orgId any) (*Organisation, error) {
	fmt.Printf("GetOrganisationById, orgId: %v, type: %T \n", orgId, orgId)
	org := Organisation{}
	err := db.GormDB.Where("id = ?", orgId).First(&org).Error
	if err != nil {
		return nil, fmt.Errorf("Error fetching organisation: %v\n", err)
	}
	return &org, nil
}

func (db *Database) CreateDiggerJob(jobId string, parentJobId *string) (*DiggerJob, error) {
	job := &DiggerJob{DiggerJobId: jobId, ParentDiggerJobId: parentJobId, Status: DiggerJobCreated}
	result := db.GormDB.Save(job)
	if result.Error != nil {
		return nil, result.Error
	}
	return job, nil
}

func (db *Database) GetPendingDiggerJobs() ([]DiggerJob, error) {
	jobs := make([]DiggerJob, 0)
	result := db.GormDB.Where("status = ? AND parent_digger_job_id is NULL ", DiggerJobCreated).Find(&jobs)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return jobs, nil
}

func (db *Database) GetDiggerJob(jobId string) (*DiggerJob, error) {
	var job *DiggerJob
	result := db.GormDB.Where("digger_job_id=? ", jobId).Find(job)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return job, nil
}

func (db *Database) GetDiggerJobsByParentId(jobId string) ([]DiggerJob, error) {
	var jobs []DiggerJob
	result := db.GormDB.Where("parent_digger_job_id=? ", jobId).Find(&jobs)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return jobs, nil
}

func (db *Database) GetOrganisation(tenantId any) (*Organisation, error) {
	var org *Organisation
	result := db.GormDB.Take(org, "external_id = ?", tenantId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return org, nil
}

func (db *Database) CreateOrganisation(name string, externalSource string, tenantId string) (*Organisation, error) {
	org := &Organisation{Name: name, ExternalSource: externalSource, ExternalId: tenantId}
	result := db.GormDB.Save(org)
	if result.Error != nil {
		fmt.Printf("Failed to create organisation: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	fmt.Printf("Organisation %s, (id: %v) has been created successfully\n", name, org.ID)
	return org, nil
}

func (db *Database) CreateProject(name string, org *Organisation, repo *Repo) (*Project, error) {
	project := &Project{Name: name, Organisation: org, Repo: repo}
	result := db.GormDB.Save(project)
	if result.Error != nil {
		fmt.Printf("Failed to create project: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	fmt.Printf("Project %s, (id: %v) has been created successfully\n", name, project.ID)
	return project, nil
}

func (db *Database) CreateRepo(name string, org *Organisation, diggerConfig string) (*Repo, error) {
	repo := &Repo{Name: name, Organisation: org, DiggerConfig: diggerConfig}
	result := db.GormDB.Save(repo)
	if result.Error != nil {
		fmt.Printf("Failed to create repo: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	fmt.Printf("Repo %s, (id: %v) has been created successfully\n", name, repo.ID)
	return repo, nil
}

func (db *Database) GetToken(tenantId any) (*Token, error) {
	var org *Token
	result := db.GormDB.Take(org, "value = ?", tenantId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return org, nil
}
