package models

import (
	"errors"
	"fmt"
	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func (db *Database) GetProjectsFromContext(c *gin.Context, orgIdKey string) ([]Project, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	log.Printf("getProjectsFromContext, org id: %v\n", loggedInOrganisationId)

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
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	log.Printf("getProjectsFromContext, number of projects:%d\n", len(projects))
	return projects, true
}

func (db *Database) GetPoliciesFromContext(c *gin.Context, orgIdKey string) ([]Policy, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	log.Printf("getPoliciesFromContext, org id: %v\n", loggedInOrganisationId)

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
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	log.Printf("getPoliciesFromContext, number of policies:%d\n", len(policies))
	return policies, true
}

func (db *Database) GetProjectRunsFromContext(c *gin.Context, orgIdKey string) ([]ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)

	log.Printf("getProjectRunsFromContext, org id: %v\n", loggedInOrganisationId)

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
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	log.Printf("getProjectRunsFromContext, number of runs:%d\n", len(runs))
	return runs, true
}

func (db *Database) GetProjectByRunId(c *gin.Context, runId uint, orgIdKey string) (*ProjectRun, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return nil, false
	}

	log.Printf("GetProjectByRunId, org id: %v\n", loggedInOrganisationId)
	var projectRun ProjectRun

	err := db.GormDB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
		Joins("INNER JOIN projects ON projects.id = project_runs.project_id").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("project_runs.id = ?", runId).First(&projectRun).Error

	if err != nil {
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
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

	log.Printf("GetProjectByProjectId, org id: %v\n", loggedInOrganisationId)
	var project Project

	err := db.GormDB.Preload("Organisation").Preload("Repo").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("projects.id = ?", projectId).First(&project).Error

	if err != nil {
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
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

	log.Printf("getPolicyByPolicyId, org id: %v\n", loggedInOrganisationId)
	var policy Policy

	err := db.GormDB.Preload("Project").Preload("Project.Organisation").Preload("Project.Repo").
		Joins("INNER JOIN projects ON projects.id = policies.project_id").
		Joins("INNER JOIN repos ON projects.repo_id = repos.id").
		Joins("INNER JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", loggedInOrganisationId).
		Where("policies.id = ?", policyId).First(&policy).Error

	if err != nil {
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &policy, true
}

func (db *Database) GetDefaultRepo(c *gin.Context, orgIdKey string) (*Repo, bool) {
	loggedInOrganisationId, exists := c.Get(orgIdKey)
	if !exists {
		log.Print("Not allowed to access this resource")
		return nil, false
	}

	log.Printf("getDefaultRepo, org id: %v\n", loggedInOrganisationId)
	var repo Repo

	err := db.GormDB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ?", loggedInOrganisationId).First(&repo).Error

	if err != nil {
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, false
	}

	return &repo, true
}

// GetRepo returns digger repo by organisationId and repo name (diggerhq-digger)
func (db *Database) GetRepo(orgIdKey any, repoName string) (*Repo, error) {
	var repo Repo

	err := db.GormDB.Preload("Organisation").
		Joins("INNER JOIN organisations ON repos.organisation_id = organisations.id").
		Where("organisations.id = ? AND repos.name=?", orgIdKey, repoName).First(&repo).Error

	if err != nil {
		log.Printf("Unknown error occurred while fetching database, %v\n", err)
		return nil, err
	}
	return &repo, nil
}

func (db *Database) GithubRepoAdded(installationId int64, appId int, login string, accountId int64, repoFullName string) error {
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
		_, err := db.CreateGithubAppInstallation(installationId, int64(appId), login, int(accountId), repoFullName)
		if err != nil {
			return fmt.Errorf("failed to save github installation item to database. %v", err)
		}
	} else {
		log.Printf("Record for installation_id: %d, repo: %s, with status=active exist already.", installationId, repoFullName)
		item.Status = GithubAppInstallActive
		item.UpdatedAt = time.Now()
		err := db.GormDB.Save(item).Error
		if err != nil {
			return fmt.Errorf("failed to update github installation in the database. %v", err)
		}
	}
	return nil
}

func (db *Database) GithubRepoRemoved(installationId int64, appId int, repoFullName string) error {
	item := GithubAppInstallation{}
	err := db.GormDB.Where("github_installation_id = ? AND status=? AND github_app_id=? AND repo=?", installationId, GithubAppInstallActive, appId, repoFullName).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Record not found for installationId: %d, status=active, githubAppId: %d and repo: %s", installationId, appId, repoFullName)
			return nil
		}
		return fmt.Errorf("failed to find github installation in database. %v", err)
	}
	item.Status = GithubAppInstallDeleted
	item.UpdatedAt = time.Now()
	err = db.GormDB.Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update github installation in the database. %v", err)
	}
	return nil
}

func (db *Database) GetGithubAppInstallationByOrgAndRepo(orgId any, repo string) (*GithubAppInstallation, error) {
	link, err := db.GetGithubInstallationLinkForOrg(orgId)
	if err != nil {
		return nil, err
	}

	installation := GithubAppInstallation{}
	result := db.GormDB.Where("github_installation_id = ? AND status=? AND repo=?", link.GithubInstallationId, GithubAppInstallationLinkActive, repo).Find(&installation)
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

// GetGithubAppInstallationByIdAndRepo repoFullName should be in the following format: org/repo_name, for example "diggerhq/github-job-scheduler"
func (db *Database) GetGithubAppInstallationByIdAndRepo(installationId int64, repoFullName string) (*GithubAppInstallation, error) {
	installation := GithubAppInstallation{}
	result := db.GormDB.Where("github_installation_id = ? AND status=? AND repo=?", installationId, GithubAppInstallActive, repoFullName).Find(&installation)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if installation.Model.ID == 0 {
		return nil, fmt.Errorf("GithubAppInstallation with id=%v doesn't exist.")
	}
	return &installation, nil
}

// GetGithubAppInstallationLink repoFullName should be in the following format: org/repo_name, for example "diggerhq/github-job-scheduler"
func (db *Database) GetGithubAppInstallationLink(installationId int64) (*GithubAppInstallationLink, error) {
	var link GithubAppInstallationLink
	result := db.GormDB.Preload("Organisation").Where("github_installation_id = ? AND status=?", installationId, GithubAppInstallationLinkActive).Find(&link)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	// If not found, the values will be default values, which means ID will be 0
	if link.Model.ID == 0 {
		return nil, nil
	}
	return &link, nil
}

// GetGithubApp
func (db *Database) GetGithubApp(gitHubAppId int64) (*GithubApp, error) {
	app := GithubApp{}
	result := db.GormDB.Where("github_id = ?", gitHubAppId).Find(&app)
	if result.Error != nil {
		return nil, result.Error
	}
	return &app, nil
}

func (db *Database) CreateGithubInstallationLink(org *Organisation, installationId int64) (*GithubAppInstallationLink, error) {
	l := GithubAppInstallationLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("github_installation_id = ? AND status=?", installationId, GithubAppInstallationLinkActive).Find(&l)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	if result.RowsAffected > 0 {
		if l.OrganisationId != org.ID {
			return nil, fmt.Errorf("GitHub app installation %v already linked to another org ", installationId)
		}
		// record already exist, do nothing
		return &l, nil
	}

	var list []GithubAppInstallationLink
	// if there are other installation for this org, we need to make them inactive
	result = db.GormDB.Preload("Organisation").Where("github_installation_id <> ? AND organisation_id = ? AND status=?", installationId, org.ID, GithubAppInstallationLinkActive).Find(&list)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	for _, item := range list {
		item.Status = GithubAppInstallationLinkInactive
		db.GormDB.Save(item)
	}

	link := GithubAppInstallationLink{Organisation: org, GithubInstallationId: installationId, Status: GithubAppInstallationLinkActive}
	result = db.GormDB.Save(&link)
	if result.Error != nil {
		return nil, result.Error
	}
	return &link, nil
}

func (db *Database) GetGithubInstallationLinkForOrg(orgId any) (*GithubAppInstallationLink, error) {
	l := GithubAppInstallationLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("organisation_id = ? AND status=?", orgId, GithubAppInstallationLinkActive).Find(&l)
	if result.Error != nil {
		return nil, result.Error
	}
	return &l, nil
}

func (db *Database) CreateDiggerJobLink(diggerJobId string, repoFullName string) (*GithubDiggerJobLink, error) {
	link := GithubDiggerJobLink{Status: DiggerJobLinkCreated, DiggerJobId: diggerJobId, RepoFullName: repoFullName}
	result := db.GormDB.Save(&link)
	if result.Error != nil {
		log.Printf("Failed to create GithubDiggerJobLink, %v, repo: %v \n", diggerJobId, repoFullName)
		return nil, result.Error
	}
	log.Printf("GithubDiggerJobLink %v, (repo: %v) has been created successfully\n", diggerJobId, repoFullName)
	return &link, nil
}

func (db *Database) UpdateDiggerJobLink(diggerJobId string, repoFullName string, githubJobId int64) (*GithubDiggerJobLink, error) {
	jobLink := GithubDiggerJobLink{}
	// check if there is already a link to another org, and throw an error in this case
	result := db.GormDB.Where("digger_job_id = ? AND repo_full_name=? ", diggerJobId, repoFullName).Find(&jobLink)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Printf("Failed to update GithubDiggerJobLink, %v, repo: %v \n", diggerJobId, repoFullName)
			return nil, result.Error
		}
	}
	if result.RowsAffected == 1 {
		jobLink.GithubJobId = githubJobId
		result = db.GormDB.Save(&jobLink)
		if result.Error != nil {
			return nil, result.Error
		}
		log.Printf("GithubDiggerJobLink %v, (repo: %v) has been updated successfully\n", diggerJobId, repoFullName)
		return &jobLink, nil
	}
	return &jobLink, nil
}

func (db *Database) GetOrganisationById(orgId any) (*Organisation, error) {
	log.Printf("GetOrganisationById, orgId: %v, type: %T \n", orgId, orgId)
	org := Organisation{}
	err := db.GormDB.Where("id = ?", orgId).First(&org).Error
	if err != nil {
		return nil, fmt.Errorf("Error fetching organisation: %v\n", err)
	}
	return &org, nil
}

func (db *Database) CreateDiggerJob(batch uuid.UUID, parentJobId *string, serializedJob []byte, branchName string) (*DiggerJob, error) {
	jobId := uniuri.New()
	job := &DiggerJob{DiggerJobId: jobId, ParentDiggerJobId: parentJobId, Status: DiggerJobCreated,
		BatchId: batch, SerializedJob: serializedJob, BranchName: branchName}
	result := db.GormDB.Save(job)
	if result.Error != nil {
		return nil, result.Error
	}

	log.Printf("DiggerJob %v, (id: %v) has been created successfully\n", job.DiggerJobId, job.ID)
	return job, nil
}

func (db *Database) UpdateDiggerJob(job *DiggerJob) error {
	result := db.GormDB.Save(job)
	if result.Error != nil {
		return result.Error
	}
	log.Printf("DiggerJob %v, (id: %v) has been updated successfully\n", job.DiggerJobId, job.ID)
	return nil
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
	var job DiggerJob
	result := db.GormDB.Where("digger_job_id=? ", jobId).Find(&job)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return &job, nil
}

func (db *Database) GetDiggerJobsByParentIdAndStatus(jobId *string, status DiggerJobStatus) ([]DiggerJob, error) {
	var jobs []DiggerJob
	result := db.GormDB.Where("parent_digger_job_id=? AND status=?", jobId, status).Find(&jobs)
	if result.Error != nil {
		log.Printf("Failed to get DiggerJob by parent job id: %v, error: %v\n", jobId, result.Error)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return jobs, nil
}

func (db *Database) GetDiggerJobsWithoutParent() ([]DiggerJob, error) {
	var jobs []DiggerJob
	result := db.GormDB.Where("parent_digger_job_id is NULL AND status=?", DiggerJobCreated).Find(&jobs)
	if result.Error != nil {
		log.Printf("Failed to Get DiggerJobsWithoutParent, error: %v\n", result.Error)
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}
	return jobs, nil
}

func (db *Database) GetOrganisation(tenantId any) (*Organisation, error) {
	var org Organisation
	result := db.GormDB.Take(&org, "external_id = ?", tenantId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return &org, nil
}

func (db *Database) CreateOrganisation(name string, externalSource string, tenantId string) (*Organisation, error) {
	org := &Organisation{Name: name, ExternalSource: externalSource, ExternalId: tenantId}
	result := db.GormDB.Save(org)
	if result.Error != nil {
		log.Printf("Failed to create organisation: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	log.Printf("Organisation %s, (id: %v) has been created successfully\n", name, org.ID)
	return org, nil
}

func (db *Database) CreateProject(name string, org *Organisation, repo *Repo) (*Project, error) {
	project := &Project{Name: name, Organisation: org, Repo: repo}
	result := db.GormDB.Save(project)
	if result.Error != nil {
		log.Printf("Failed to create project: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	log.Printf("Project %s, (id: %v) has been created successfully\n", name, project.ID)
	return project, nil
}

func (db *Database) CreateRepo(name string, org *Organisation, diggerConfig string) (*Repo, error) {
	repo := &Repo{Name: name, Organisation: org, DiggerConfig: diggerConfig}
	result := db.GormDB.Save(repo)
	if result.Error != nil {
		log.Printf("Failed to create repo: %v, error: %v\n", name, result.Error)
		return nil, result.Error
	}
	log.Printf("Repo %s, (id: %v) has been created successfully\n", name, repo.ID)
	return repo, nil
}

func (db *Database) GetToken(tenantId any) (*Token, error) {
	var token Token
	result := db.GormDB.Take(&token, "value = ?", tenantId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, result.Error
		}
	}
	return &token, nil
}

func (db *Database) CreateGithubAppInstallation(installationId int64, githubAppId int64, login string, accountId int, repoFullName string) (*GithubAppInstallation, error) {
	installation := &GithubAppInstallation{
		GithubInstallationId: installationId,
		GithubAppId:          githubAppId,
		Login:                login,
		AccountId:            accountId,
		Repo:                 repoFullName,
		Status:               GithubAppInstallActive,
	}
	result := db.GormDB.Save(installation)
	if result.Error != nil {
		log.Printf("Failed to create GithubAppInstallation: %v, error: %v\n", installationId, result.Error)
		return nil, result.Error
	}
	log.Printf("GithubAppInstallation %v, (id: %v) has been created successfully\n", installationId, installation.ID)
	return installation, nil
}
