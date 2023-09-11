package controllers

import (
	"context"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"digger.dev/cloud/services"
	"digger.dev/cloud/utils"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dchest/uniuri"
	dg_configuration "github.com/diggerhq/lib-digger-config"
	orchestrator "github.com/diggerhq/lib-orchestrator"
	dg_github "github.com/diggerhq/lib-orchestrator/github"
	dg_github_models "github.com/diggerhq/lib-orchestrator/github/models"
	webhooks "github.com/diggerhq/webhooks/github"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

func GithubAppWebHook(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	gh := &utils.DiggerGithubRealClient{}

	_, err := github.ValidatePayload(c.Request, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
	if err != nil {
		log.Printf("Error validating github app webhook's payload: %v", err)
		c.String(http.StatusBadRequest, "Error validating github app webhook's payload")
		return
	}

	hook, _ := webhooks.New()

	payload, err := hook.Parse(c.Request, webhooks.InstallationEvent, webhooks.PullRequestEvent, webhooks.IssueCommentEvent,
		webhooks.InstallationRepositoriesEvent, webhooks.WorkflowJobEvent, webhooks.WorkflowRunEvent)
	if err != nil {
		if errors.Is(err, webhooks.ErrEventNotFound) {
			// ok event wasn't one of the ones asked to be parsed
			fmt.Println("Github event  wasn't found.")
		}
		fmt.Printf("Failed to parse Github Event. :%v\n", err)
		c.String(http.StatusInternalServerError, "Failed to parse Github Event")
		return
	}
	switch payload.(type) {

	case webhooks.InstallationPayload:
		installation := payload.(webhooks.InstallationPayload)
		if installation.Action == "created" {
			err := handleInstallationCreatedEvent(installation)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle webhook event.")
				return
			}
		}

		if installation.Action == "deleted" {
			err := handleInstallationDeletedEvent(installation)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle webhook event.")
				return
			}

		}
	case webhooks.InstallationRepositoriesPayload:
		installationRepos := payload.(webhooks.InstallationRepositoriesPayload)
		if installationRepos.Action == "added" {
			installationId := installationRepos.Installation.ID
			login := installationRepos.Installation.Account.Login
			accountId := installationRepos.Installation.Account.ID
			appId := installationRepos.Installation.AppID
			for _, repo := range installationRepos.RepositoriesAdded {
				err := models.DB.GithubRepoAdded(installationId, appId, login, accountId, repo.FullName)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to store item.")
					return
				}
			}
		}
		if installationRepos.Action == "removed" {
			installationId := installationRepos.Installation.ID
			appId := installationRepos.Installation.AppID
			for _, repo := range installationRepos.RepositoriesRemoved {
				err := models.DB.GithubRepoRemoved(installationId, appId, repo.FullName)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to remove item.")
					return
				}
			}
		}
	case webhooks.IssueCommentPayload:
		payload := payload.(webhooks.IssueCommentPayload)
		err := handleIssueCommentEvent(gh, &payload)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	case webhooks.WorkflowJobPayload:
		payload := payload.(webhooks.WorkflowJobPayload)
		err := handleWorkflowJobEvent(gh, payload)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to handle WorkflowJob event.")
			return
		}
	case webhooks.WorkflowRunPayload:
		payload := payload.(webhooks.WorkflowRunPayload)
		err := handleWorkflowRunEvent(payload)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to handle WorkflowRun event.")
			return
		}
	case webhooks.PullRequestPayload:
		payload := payload.(webhooks.PullRequestPayload)
		log.Printf("Got pull request event for %v", payload.PullRequest.ID)
		err := handlePullRequestEvent(gh, &payload)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(200, "ok")
}

func handleInstallationCreatedEvent(installation webhooks.InstallationPayload) error {
	installationId := installation.Installation.ID
	login := installation.Installation.Account.Login
	accountId := installation.Installation.Account.ID
	appId := installation.Installation.AppID

	for _, repo := range installation.Repositories {
		fmt.Printf("Adding a new installation %d for repo: %s", installationId, repo.FullName)
		err := models.DB.GithubRepoAdded(installationId, appId, login, accountId, repo.FullName)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleInstallationDeletedEvent(installation webhooks.InstallationPayload) error {
	installationId := installation.Installation.ID
	appId := installation.Installation.AppID
	for _, repo := range installation.Repositories {
		fmt.Printf("Removing an installation %d for repo: %s", installationId, repo.FullName)
		err := models.DB.GithubRepoRemoved(installationId, appId, repo.FullName)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleWorkflowJobEvent(gh utils.DiggerGithubClient, payload webhooks.WorkflowJobPayload) error {
	ctx := context.Background()
	switch payload.Action {
	case "completed":
		githubJobId := payload.WorkflowJob.ID
		//githubJobStatus := payload.WorkflowJob.Status

		repo := payload.Repository.Name
		owner := payload.Repository.Owner.Login
		repoFullName := payload.Repository.FullName
		installationId := payload.Installation.ID

		installation, err := models.DB.GetGithubAppInstallationByIdAndRepo(installationId, repoFullName)
		if err != nil {
			return err
		}
		client, _, err := gh.GetGithubClient(installation.GithubAppId, installationId)
		if err != nil {
			return err
		}

		workflowJob, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, githubJobId)
		if err != nil {
			return err
		}

		var jobId string
		for _, s := range (*workflowJob).Steps {
			name := *s.Name
			if strings.HasPrefix(name, "digger run ") {
				// digger job id and workflow step name matched
				jobId = strings.Replace(name, "digger run ", "", 1)
				_, err := models.DB.UpdateDiggerJobLink(repoFullName, jobId, githubJobId)
				if err != nil {
					return err
				}
			}
		}
		if jobId != "" {
			workflowFileName := "workflow.yml"
			job, err := models.DB.GetDiggerJob(jobId)
			if err != nil {
				return err
			}
			err = services.DiggerJobCompleted(client, job, owner, repo, workflowFileName)
			if err != nil {
				return err
			}
		}

	case "queued":
	case "in_progress":
	}
	return nil
}

func handleWorkflowRunEvent(payload webhooks.WorkflowRunPayload) error {
	return nil
}

func handlePullRequestEvent(gh utils.DiggerGithubClient, payload *webhooks.PullRequestPayload) error {
	var installationId int64
	var repoName string
	var repoOwner string
	var repoFullName string
	var cloneURL string
	var actor string

	installationId = payload.Installation.ID
	repoName = payload.Repository.Name
	repoOwner = payload.Repository.Owner.Login
	repoFullName = payload.Repository.FullName
	cloneURL = payload.Repository.CloneURL
	actor = payload.Sender.Login

	installation, err := models.DB.GetGithubAppInstallationByIdAndRepo(installationId, repoFullName)
	if err != nil {
		log.Printf("Error getting installation: %v", err)
		return fmt.Errorf("error getting github app installation")
	}

	_, err = models.DB.GetGithubApp(installation.GithubAppId)
	if err != nil {
		log.Printf("Error getting app: %v", err)
		return fmt.Errorf("error getting github app")
	}

	ghClient, token, err := gh.GetGithubClient(installation.GithubAppId, installation.GithubInstallationId)
	if err != nil {
		log.Printf("Error creating github app client: %v", err)
		return fmt.Errorf("error creating github app client")
	}

	ghService := dg_github.GithubService{
		Client:   ghClient,
		RepoName: repoName,
		Owner:    repoOwner,
	}

	prBranch := payload.PullRequest.Head.Ref

	link, err := models.DB.GetGithubAppInstallationLinkByIdAndRepo(installationId, repoFullName)
	if err != nil {
		log.Printf("Error getting GithubAppInstallationLink: %v", err)
		return fmt.Errorf("error getting github app link to installation")
	}

	diggerRepoName := repoOwner + "-" + repoName
	repo, ok := models.DB.GetRepo(link.Organisation.ID, diggerRepoName)
	if !ok {
		log.Printf("Error getting repo: %v", err)
		return fmt.Errorf("error getting repo")
	}

	configYaml, err := dg_configuration.LoadDiggerConfigYamlFromString(repo.DiggerConfig)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return fmt.Errorf("error loading digger config")
	}

	if configYaml.GenerateProjectsConfig != nil {
		err = utils.CloneGitRepoAndDoAction(cloneURL, prBranch, *token, func(dir string) {
			dg_configuration.HandleYamlProjectGeneration(configYaml, dir)
		})
		if err != nil {
			log.Printf("Error generating projects: %v", err)
			return fmt.Errorf("error generating projects")
		}
	}

	config, _, err := loadDiggerConfig(configYaml)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return fmt.Errorf("error loading digger config")
	}

	impactedProjects, requestedProject, prNumber, err := dg_github.ProcessGitHubEvent(payload, config, &ghService)

	if err != nil {
		log.Printf("Error processing event: %v", err)
		return fmt.Errorf("error processing event")
	}
	eventPackage := dg_github_models.EventPackage{
		Event:      payload,
		EventName:  "pull_request",
		Actor:      actor,
		Repository: repoFullName,
	}

	jobs, _, err := dg_github.ConvertGithubEventToJobs(eventPackage, impactedProjects, requestedProject, config.Workflows)

	if err != nil {
		log.Printf("Error converting event to jobs: %v", err)
		return fmt.Errorf("error converting event to jobs")
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

			_, err = ghClient.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), repoOwner, repoName, "workflow.yml", github.CreateWorkflowDispatchEventRequest{
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
	wg.Wait()
	for projecName, success := range successPerJob {
		err := ghService.PublishComment(prNumber, fmt.Sprintf("Digger has %v the %v project", map[bool]string{true: "started", false: "failed to start"}[success], projecName))
		if err != nil {
			log.Printf("Error publishing comment: %v", err)
		}
	}
	return nil
}

func handleIssueCommentEvent(gh utils.DiggerGithubClient, payload *webhooks.IssueCommentPayload) error {
	var installationId int64
	var repoName string
	var repoOwner string
	var repoFullName string
	var cloneURL string
	var actor string

	installationId = payload.Installation.ID
	repoName = payload.Repository.Name
	repoOwner = payload.Repository.Owner.Login
	repoFullName = payload.Repository.FullName
	cloneURL = payload.Repository.CloneURL
	actor = payload.Sender.Login

	installation, err := models.DB.GetGithubAppInstallationByIdAndRepo(installationId, repoFullName)
	if err != nil {
		log.Printf("Error getting installation: %v", err)
		return fmt.Errorf("error getting installation")
	}

	_, err = models.DB.GetGithubApp(installation.GithubAppId)
	if err != nil {
		log.Printf("Error getting app: %v", err)
		return fmt.Errorf("error getting app")
	}

	ghClient, token, err := gh.GetGithubClient(installation.GithubAppId, installation.GithubInstallationId)
	if err != nil {
		log.Printf("Error creating github app client: %v", err)
		return fmt.Errorf("error creating github app client")
	}

	ghService := dg_github.GithubService{
		Client:   ghClient,
		RepoName: repoName,
		Owner:    repoOwner,
	}

	var prBranch string
	prBranch, err = ghService.GetBranchName(int(payload.Issue.Number))
	if err != nil {
		log.Printf("Error getting branch name: %v", err)
		return fmt.Errorf("error getting branch name")
	}

	link, err := models.DB.GetGithubAppInstallationLinkByIdAndRepo(installationId, repoFullName)
	if err != nil {
		log.Printf("Error getting branch name: %v", err)
		return fmt.Errorf("error getting branch name")
	}

	diggerRepoName := repoOwner + "-" + repoName
	repo, ok := models.DB.GetRepo(link.Organisation.ID, diggerRepoName)
	if !ok {
		log.Printf("Error getting repo: %v", err)
		return fmt.Errorf("error getting repo")
	}

	configYaml, err := dg_configuration.LoadDiggerConfigYamlFromString(repo.DiggerConfig)
	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return fmt.Errorf("error loading digger config")
	}

	if configYaml.GenerateProjectsConfig != nil {
		err = utils.CloneGitRepoAndDoAction(cloneURL, prBranch, *token, func(dir string) {
			dg_configuration.HandleYamlProjectGeneration(configYaml, dir)
		})
		if err != nil {
			log.Printf("Error generating projects: %v", err)
			return fmt.Errorf("error generating projects")
		}
	}

	config, _, err := loadDiggerConfig(configYaml)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return fmt.Errorf("error loading digger config")
	}

	impactedProjects, requestedProject, prNumber, err := dg_github.ProcessGitHubEvent(payload, config, &ghService)

	if err != nil {
		log.Printf("Error processing event: %v", err)
		return fmt.Errorf("error processing event")
	}
	eventPackage := dg_github_models.EventPackage{
		Event:      payload,
		EventName:  "pull_request",
		Actor:      actor,
		Repository: repoFullName,
	}

	jobs, _, err := dg_github.ConvertGithubEventToJobs(eventPackage, impactedProjects, requestedProject, config.Workflows)

	if err != nil {
		log.Printf("Error converting event to jobs: %v", err)
		return fmt.Errorf("error converting event to jobs")
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

			_, err = ghClient.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), repoOwner, repoName, "workflow.yml", github.CreateWorkflowDispatchEventRequest{
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

	wg.Wait()
	for projecName, success := range successPerJob {
		err := ghService.PublishComment(prNumber, fmt.Sprintf("Digger has %v the %v project", map[bool]string{true: "started", false: "failed to start"}[success], projecName))
		if err != nil {
			log.Printf("Error publishing comment: %v", err)
		}
	}
	return nil
}

func GithubAppCallbackPage(c *gin.Context) {
	installationId := c.Request.URL.Query()["installation_id"][0]
	//setupAction := c.Request.URL.Query()["setup_action"][0]
	code := c.Request.URL.Query()["code"][0]
	clientId := os.Getenv("GITHUB_APP_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_APP_CLIENT_SECRET")

	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	installationId64, err := strconv.ParseInt(installationId, 10, 64)
	if err != nil {
		fmt.Printf("err: %v", err)
		c.String(http.StatusInternalServerError, "Failed to parse installation_id.")
		return
	}

	result, err := validateGithubCallback(clientId, clientSecret, code, installationId64)
	if !result {
		fmt.Printf("Failed to validated installation id, %v\n", err)
		c.String(http.StatusInternalServerError, "Failed to validate installation_id.")
		return
	}

	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	_, err = models.DB.CreateGithubInstallationLink(org.ID, installationId64)
	if err != nil {
		log.Printf("Error saving CreateGithubInstallationLink to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating GitHub installation"})
		return
	}
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
}

func GihHubCreateTestJobPage(c *gin.Context) {
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		log.Printf("Organisation ID not found in context")
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	diggerJobId := uniuri.New()
	parentJobId := uniuri.New()
	job, err := models.DB.CreateDiggerJob(parentJobId, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating digger job"})
		return
	}
	_, err = models.DB.CreateDiggerJob(diggerJobId, &parentJobId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating digger job"})
		return
	}
	/*
		jobs, err := models.GetPendingDiggerJobs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating digger job"})
			return
		}

		for _, j := range jobs {
			fmt.Printf("jobId: %v, parentJobId: %v", j.DiggerJobId, j.ParentDiggerJobId)
		}
	*/

	owner := "diggerhq"
	repo := "github-job-scheduler"
	workflowFileName := "workflow.yml"
	repoFullName := owner + "/" + repo

	installation, err := models.DB.GetGithubAppInstallationByOrgAndRepo(orgId, repoFullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating github installation"})
		return
	}

	/*
		link, err := models.CreateDiggerJobLink(repoFullName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a test job"})
			return
		}
	*/
	gh := &utils.DiggerGithubRealClient{}
	client, _, err := gh.GetGithubClient(installation.GithubAppId, installation.GithubInstallationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating a token"})
		return
	}

	services.TriggerTestJob(client, owner, repo, job, workflowFileName)
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
}

// why this validation is needed: https://roadie.io/blog/avoid-leaking-github-org-data/
// validation based on https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app , step 3
func validateGithubCallback(clientId string, clientSecret string, code string, installationId int64) (bool, error) {
	ctx := context.Background()
	type OAuthAccessResponse struct {
		AccessToken string `json:"access_token"`
	}
	httpClient := http.Client{}

	reqURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientId, clientSecret, code)
	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return false, fmt.Errorf("could not create HTTP request: %v\n", err)
	}
	req.Header.Set("accept", "application/json")

	res, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request to login/oauth/access_token failed: %v\n", err)
	}

	if err != nil {
		return false, fmt.Errorf("Failed to read response's body: %v\n", err)
	}

	var t OAuthAccessResponse
	if err := json.NewDecoder(res.Body).Decode(&t); err != nil {
		return false, fmt.Errorf("could not parse JSON response: %v\n", err)
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: t.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	installationIdMatch := false
	// list all installations for the user
	installations, _, err := client.Apps.ListUserInstallations(ctx, nil)
	for _, v := range installations {
		fmt.Printf("installation id: %v\n", *v.ID)
		if *v.ID == installationId {
			installationIdMatch = true
		}
	}
	if !installationIdMatch {
		return false, fmt.Errorf("InstallationId %v doesn't match any id for specified user\n", installationId)
	}
	return true, nil
}
