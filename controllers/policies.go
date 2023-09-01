package controllers

import (
	"context"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"digger.dev/cloud/utils"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	dg_configuration "github.com/diggerhq/lib-digger-config"
	orchestrator "github.com/diggerhq/lib-orchestrator"
	dg_github "github.com/diggerhq/lib-orchestrator/github"
	dg_github_models "github.com/diggerhq/lib-orchestrator/github/models"
	"github.com/dominikbraun/graph"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v53/github"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type CreatePolicyInput struct {
	Policy string
}

func FindAccessPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_ACCESS)
}

func FindPlanPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_PLAN)
}

func FindDriftPolicy(c *gin.Context) {
	findPolicy(c, models.POLICY_TYPE_DRIFT)
}

func findPolicy(c *gin.Context, policyType string) {
	repo := c.Param("repo")
	projectName := c.Param("projectName")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		log.Printf("Organisation ID not found in context")
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var policy models.Policy
	query := JoinedOrganisationRepoProjectQuery()

	if repo != "" && projectName != "" {
		err := query.
			Where("repos.name = ? AND projects.name = ? AND policies.organisation_id = ? AND policies.type = ?", repo, projectName, orgId, policyType).
			First(&policy).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, fmt.Sprintf("Could not find policy for repo %v and project name %v", repo, projectName))
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else {
		c.String(http.StatusBadRequest, "Should pass repo and project name")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

func FindAccessPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_ACCESS)
}

func FindPlanPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_PLAN)
}

func FindDriftPolicyForOrg(c *gin.Context) {
	findPolicyForOrg(c, models.POLICY_TYPE_DRIFT)
}

func findPolicyForOrg(c *gin.Context, policyType string) {
	organisation := c.Param("organisation")
	var policy models.Policy
	query := JoinedOrganisationRepoProjectQuery()

	err := query.
		Where("organisations.name = ? AND (repos.id IS NULL AND projects.id IS NULL) AND policies.type = ? ", organisation, policyType).
		First(&policy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Could not find policy for organisation: "+organisation)
		} else {
			c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		}
		return
	}

	loggedInOrganisation := c.GetUint(middleware.ORGANISATION_ID_KEY)

	if policy.OrganisationID != loggedInOrganisation {
		log.Printf("Organisation ID %v does not match logged in organisation ID %v", policy.OrganisationID, loggedInOrganisation)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

func JoinedOrganisationRepoProjectQuery() *gorm.DB {
	return models.DB.Preload("Organisation").Preload("Repo").Preload("Project").
		Joins("LEFT JOIN repos ON policies.repo_id = repos.id").
		Joins("LEFT JOIN projects ON policies.project_id = projects.id").
		Joins("LEFT JOIN organisations ON policies.organisation_id = organisations.id")
}

func UpsertAccessPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_ACCESS)
}

func UpsertPlanPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_PLAN)
}

func UpsertDriftPolicyForOrg(c *gin.Context) {
	upsertPolicyForOrg(c, models.POLICY_TYPE_DRIFT)
}

func upsertPolicyForOrg(c *gin.Context, policyType string) {
	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	organisation := c.Param("organisation")

	org := models.Organisation{}
	orgResult := models.DB.Where("name = ?", organisation).Take(&org)
	if orgResult.RowsAffected == 0 {
		c.String(http.StatusNotFound, "Could not find organisation: "+organisation)
		return
	}

	loggedInOrganisation := c.GetUint(middleware.ORGANISATION_ID_KEY)

	if org.ID != loggedInOrganisation {
		log.Printf("Organisation ID %v does not match logged in organisation ID %v", org.ID, loggedInOrganisation)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	policy := models.Policy{}

	policyResult := models.DB.Where("organisation_id = ? AND (repo_id IS NULL AND project_id IS NULL) AND type = ?", org.ID, policyType).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: org.ID,
			Type:           policyType,
			Policy:         string(policyData),
		}).Error

		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func UpsertAccessPolicyForRepoAndProject(c *gin.Context) {
	upsertPolicyForRepoAndProject(c, models.POLICY_TYPE_ACCESS)
}

func UpsertPlanPolicyForRepoAndProject(c *gin.Context) {
	upsertPolicyForRepoAndProject(c, models.POLICY_TYPE_PLAN)
}

func UpsertDriftPolicyForRepoAndProject(c *gin.Context) {
	upsertPolicyForRepoAndProject(c, models.POLICY_TYPE_DRIFT)
}

func upsertPolicyForRepoAndProject(c *gin.Context, policyType string) {
	orgID, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	orgID = orgID.(uint)

	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	repo := c.Param("repo")
	projectName := c.Param("projectName")
	repoModel := models.Repo{}
	repoResult := models.DB.Where("name = ?", repo).Take(&repoModel)
	if repoResult.RowsAffected == 0 {
		repoModel = models.Repo{
			OrganisationID: orgID.(uint),
			Name:           repo,
		}
		result := models.DB.Create(&repoModel)
		if result.Error != nil {
			log.Printf("Error creating repo: %v", err)
			c.String(http.StatusInternalServerError, "Error creating missing repo")
			return
		}
	}

	projectModel := models.Project{}
	projectResult := models.DB.Where("name = ?", projectName).Take(&projectModel)
	if projectResult.RowsAffected == 0 {
		projectModel = models.Project{
			OrganisationID: orgID.(uint),
			RepoID:         repoModel.ID,
			Name:           projectName,
		}
		err := models.DB.Create(&projectModel).Error
		if err != nil {
			log.Printf("Error creating project: %v", err)
			c.String(http.StatusInternalServerError, "Error creating missing project")
			return
		}
	}

	var policy models.Policy

	policyResult := models.DB.Where("organisation_id = ? AND repo_id = ? AND project_id = ? AND type = ?", orgID, repoModel.ID, projectModel.ID, policyType).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: orgID.(uint),
			RepoID:         &repoModel.ID,
			ProjectID:      &projectModel.ID,
			Type:           policyType,
			Policy:         string(policyData),
		}).Error
		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func IssueAccessTokenForOrg(c *gin.Context) {
	organisation_ID, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	org := models.Organisation{}
	orgResult := models.DB.Where("id = ?", organisation_ID).Take(&org)
	if orgResult.RowsAffected == 0 {
		log.Printf("Could not find organisation: %v", organisation_ID)
		c.String(http.StatusInternalServerError, "Unexpected error")
		return
	}

	// prefixing token to make easier to retire this type of tokens later
	token := "t:" + uuid.New().String()

	err := models.DB.Create(&models.Token{
		Value:          token,
		OrganisationID: org.ID,
		Type:           models.AccessPolicyType,
	}).Error

	if err != nil {
		log.Printf("Error creating token: %v", err)
		c.String(http.StatusInternalServerError, "Unexpected error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func GithubWebhookHandler(c *gin.Context) {
	payload, err := github.ValidatePayload(c.Request, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Printf("Error validating payload: %v", err)
		c.String(http.StatusBadRequest, "Error validating payload")
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request), payload)
	if err != nil {
		log.Printf("Error parsing webhook: %v", err)
		c.String(http.StatusBadRequest, "Error parsing webhook")
		return
	}

	switch event := event.(type) {
	case *github.InstallationEvent:
		log.Printf("Got installation event for %v", event.GetInstallation().GetAccount().GetLogin())
		if event.GetAction() == "created" {
			err := models.DB.Create(&models.GithubAppInstallation{
				GithubInstallationId: *event.Installation.ID,
				GithubAppId:          *event.Installation.AppID,
			}).Error
			if err != nil {
				log.Printf("Error creating github event: %v", err)
			}
			c.String(http.StatusOK, "OK")
		}
	case *github.InstallationRepositoriesEvent:
		log.Printf("Got installation event for %v", event.GetInstallation().GetAccount().GetLogin())
		if event.GetAction() == "added" {
			err := models.DB.Create(&models.GithubAppInstallation{
				GithubInstallationId: *event.Installation.ID,
				GithubAppId:          *event.Installation.AppID,
			}).Error
			if err != nil {
				log.Printf("Error creating github event: %v", err)
			}
			c.String(http.StatusOK, "OK")
		}
	case *github.PullRequestEvent:
		log.Printf("Got pull request event for %v", event.GetPullRequest().GetTitle())
		handlePullRequestRelatedEvent(c, *event)
	case *github.IssueCommentEvent:
		if event.Sender.Type != nil && *event.Sender.Type == "Bot" {
			c.String(http.StatusOK, "OK")
			return
		}
		handlePullRequestRelatedEvent(c, *event)
	default:
		log.Printf("Unhandled event type: %v", event)
		c.String(http.StatusBadRequest, "Unhandled event type")
		return
	}
}

func handlePullRequestRelatedEvent(c *gin.Context, event interface{}) {
	var installationId int64
	var repoName string
	var repoOwner string
	var repoFullName string
	var cloneURL string
	var actor string

	switch event := event.(type) {
	case github.PullRequestEvent:
		installationId = *event.Installation.ID
		repoName = *event.Repo.Name
		repoOwner = *event.Repo.Owner.Login
		repoFullName = *event.Repo.FullName
		cloneURL = *event.Repo.CloneURL
		actor = *event.Sender.Login
	case github.IssueCommentEvent:
		installationId = *event.Installation.ID
		repoName = *event.Repo.Name
		repoOwner = *event.Repo.Owner.Login
		repoFullName = *event.Repo.FullName
		cloneURL = *event.Repo.CloneURL
		actor = *event.Sender.Login
	default:
		log.Printf("Unhandled event type: %T", event)
		c.String(http.StatusInternalServerError, "Error getting installation")
		return
	}

	installation := models.GithubAppInstallation{}
	err := models.DB.Where("github_installation_id = ?", installationId).Take(&installation).Error
	if err != nil {
		log.Printf("Error getting installation: %v", err)
		c.String(http.StatusInternalServerError, "Error getting installation")
		return
	}
	ghApp := models.GithubApp{}
	err = models.DB.Where("github_id = ?", installation.GithubAppId).Take(&ghApp).Error
	if err != nil {
		log.Printf("Error getting app: %v", err)
		c.String(http.StatusInternalServerError, "Error getting app")
		return
	}
	tr := http.DefaultTransport

	itr, err := ghinstallation.New(tr, installation.GithubAppId, installation.GithubInstallationId, []byte(ghApp.PrivateKey))
	if err != nil {
		log.Printf("Error initialising installation: %v", err)
		c.String(http.StatusInternalServerError, "Error getting app")
		return
	}

	ghClient := github.NewClient(&http.Client{Transport: itr})

	ghService := dg_github.GithubService{
		Client:   ghClient,
		RepoName: repoName,
		Owner:    repoOwner,
	}

	var prBranch string

	switch event := event.(type) {
	case github.PullRequestEvent:
		prBranch = event.PullRequest.Head.GetRef()
	case github.IssueCommentEvent:
		prBranch, err = ghService.GetBranchName(event.Issue.GetNumber())
		if err != nil {
			log.Printf("Error getting branch name: %v", err)
			c.String(http.StatusInternalServerError, "Error getting branch name")
			return
		}
	default:
		log.Printf("Unhandled event type: %T", event)
		c.String(http.StatusInternalServerError, "Error getting branch name")
		return
	}

	var repo models.Repo

	err = models.DB.Where("name = ? AND organisation_id = ?", strings.ReplaceAll(repoFullName, "/", "-"), ghApp.OrganisationId).Take(&repo).Error

	if err != nil {
		log.Printf("Error getting repo: %v", err)
		c.String(http.StatusInternalServerError, "Error getting repo")
		return
	}

	configYaml, err := dg_configuration.LoadDiggerConfigYamlFromString(repo.DiggerConfig)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		c.String(http.StatusInternalServerError, "Error loading digger config")
		return
	}

	if configYaml.GenerateProjectsConfig != nil {
		token, err := itr.Token(context.Background())
		if err != nil {
			log.Printf("Error getting token: %v", err)
			c.String(http.StatusInternalServerError, "Error getting token")
			return
		}
		err = utils.CloneGitRepoAndDoAction(cloneURL, prBranch, token, func(dir string) {
			dg_configuration.HandleYamlProjectGeneration(configYaml, dir)
		})
		if err != nil {
			log.Printf("Error generating projects: %v", err)
			c.String(http.StatusInternalServerError, "Error generating projects")
			return
		}
	}

	config, _, err := loadDiggerConfig(configYaml)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		c.String(http.StatusInternalServerError, "Error loading digger config")
		return
	}

	impactedProjects, requestedProject, prNumber, err := dg_github.ProcessGitHubEvent(event, config, &ghService)

	if err != nil {
		log.Printf("Error processing event: %v", err)
		c.String(http.StatusInternalServerError, "Error processing event")
		return
	}
	eventPackage := dg_github_models.EventPackage{
		Event:      event,
		EventName:  "pull_request",
		Actor:      actor,
		Repository: repoFullName,
	}

	jobs, _, err := dg_github.ConvertGithubEventToJobs(eventPackage, impactedProjects, requestedProject, config.Workflows)

	if err != nil {
		log.Printf("Error converting event to jobs: %v", err)
		c.String(http.StatusInternalServerError, "Error converting event to jobs")
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(jobs))
	successPerJob := make(map[string]bool, len(jobs))

	for _, job := range jobs {
		go func(job orchestrator.Job) {
			defer wg.Done()

			marshalled, err := json.Marshal(orchestrator.JobToJson(job))

			if err != nil {
				successPerJob[job.ProjectName] = false
				log.Printf("Error marshalling job: %v", err)
				return
			}

			_, err = ghClient.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), repoOwner, repoName, "plan.yml", github.CreateWorkflowDispatchEventRequest{
				Ref:    prBranch,
				Inputs: map[string]interface{}{"job": string(marshalled)},
			})
			if err != nil {
				successPerJob[job.ProjectName] = false
				log.Printf("Error dispatching workflow: %v", err)
				return
			}
			successPerJob[job.ProjectName] = true
		}(job)
	}
	c.String(http.StatusOK, "OK")
	wg.Wait()
	for projecName, success := range successPerJob {
		err := ghService.PublishComment(prNumber, fmt.Sprintf("Digger has %v the %v project", map[bool]string{true: "started", false: "failed to start"}[success], projecName))
		if err != nil {
			log.Printf("Error publishing comment: %v", err)
		}
	}
	return
}

func loadDiggerConfig(configYaml *dg_configuration.DiggerConfigYaml) (*dg_configuration.DiggerConfig, graph.Graph[string, string], error) {

	err := dg_configuration.ValidateDiggerConfigYaml(configYaml, "loaded config")

	if err != nil {
		return nil, nil, fmt.Errorf("error validating config: %v", err)
	}

	config, depGraph, err := dg_configuration.ConvertDiggerYamlToConfig(configYaml)

	if err != nil {
		return nil, nil, fmt.Errorf("error converting config: %v", err)
	}

	err = dg_configuration.ValidateDiggerConfig(config)

	if err != nil {
		return nil, nil, fmt.Errorf("error validating config: %v", err)
	}
	return config, depGraph, nil
}
