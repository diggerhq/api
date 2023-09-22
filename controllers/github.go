package controllers

import (
	"context"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"digger.dev/cloud/utils"
	"encoding/json"
	"errors"
	"fmt"
	dg_configuration "github.com/diggerhq/lib-digger-config"
	orchestrator "github.com/diggerhq/lib-orchestrator"
	dg_github "github.com/diggerhq/lib-orchestrator/github"
	webhooks "github.com/diggerhq/webhooks/github"
	"github.com/dominikbraun/graph"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v55/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GithubAppWebHook(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	gh := &utils.DiggerGithubRealClientProvider{}

	// TODO return validation back
	/*
		_, err := github.ValidatePayload(c.Request, []byte(os.Getenv("GITHUB_WEBHOOK_SECRET")))
		if err != nil {
			log.Printf("Error validating github app webhook's payload: %v", err)
			c.String(http.StatusBadRequest, "Error validating github app webhook's payload")
			return
		}
	*/

	hook, _ := webhooks.New()

	payload, err := hook.Parse(c.Request, webhooks.InstallationEvent, webhooks.PullRequestEvent, webhooks.IssueCommentEvent,
		webhooks.InstallationRepositoriesEvent, webhooks.WorkflowJobEvent, webhooks.WorkflowRunEvent)
	if err != nil {
		if errors.Is(err, webhooks.ErrEventNotFound) {
			// ok event wasn't one of the ones asked to be parsed
			log.Println("Github event  wasn't found.")
		}
		log.Printf("Failed to parse Github Event. :%v\n", err)
		c.String(http.StatusInternalServerError, "Failed to parse Github Event")
		return
	}
	switch payload.(type) {

	case webhooks.InstallationPayload:
		payload := payload.(webhooks.InstallationPayload)
		if payload.Action == "created" {
			err := handleInstallationCreatedEvent(&payload)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle webhook event.")
				return
			}
		}

		if payload.Action == "deleted" {
			err := handleInstallationDeletedEvent(&payload)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle webhook event.")
				return
			}

		}
	case webhooks.InstallationRepositoriesPayload:
		payload := payload.(webhooks.InstallationRepositoriesPayload)
		if payload.Action == "added" {
			err := handleInstallationRepositoriesAddedEvent(&payload)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle installation repo added event.")
			}
		}
		if payload.Action == "removed" {
			err := handleInstallationRepositoriesDeletedEvent(&payload)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to handle installation repo deleted event.")
			}
		}
	case webhooks.IssueCommentPayload:
		payload := payload.(webhooks.IssueCommentPayload)
		err := handleIssueCommentEvent(gh, &payload)
		if err != nil {
			log.Printf("handleIssueCommentEvent error: %v", err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	case webhooks.PullRequestPayload:
		payload := payload.(webhooks.PullRequestPayload)
		log.Printf("Got pull request event for %v", payload.PullRequest.ID)
		err := handlePullRequestEvent(gh, &payload)
		if err != nil {
			log.Printf("handlePullRequestEvent error: %v", err)
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	c.JSON(200, "ok")
}

func createDiggerRepoForGithubRepo(ghRepoFullName string, installationId int64) error {
	link, err := models.DB.GetGithubInstallationLinkForInstallationId(installationId)
	if err != nil {
		log.Printf("Error fetching installation link: %v", err)
		return err
	}
	orgId := link.OrganisationId
	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		log.Printf("Error fetching organisation by id: %v, error: %v\n", orgId, err)
		return err
	}

	diggerRepoName := strings.ReplaceAll(ghRepoFullName, "/", "-")
	repo, err := models.DB.CreateRepo(diggerRepoName, org, `
generate_projects:
 include: "."
`)
	if err != nil {
		log.Printf("Error creating digger repo: %v", err)
		return err
	}
	log.Printf("Created digger repo: %v", repo)
	return nil
}

func handleInstallationRepositoriesAddedEvent(payload *webhooks.InstallationRepositoriesPayload) error {
	installationId := payload.Installation.ID
	login := payload.Installation.Account.Login
	accountId := payload.Installation.Account.ID
	appId := payload.Installation.AppID
	for _, repo := range payload.RepositoriesAdded {
		err := models.DB.GithubRepoAdded(installationId, appId, login, accountId, repo.FullName)
		if err != nil {
			return err
		}

		err = createDiggerRepoForGithubRepo(repo.FullName, installationId)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleInstallationRepositoriesDeletedEvent(payload *webhooks.InstallationRepositoriesPayload) error {
	installationId := payload.Installation.ID
	appId := payload.Installation.AppID
	for _, repo := range payload.RepositoriesRemoved {
		err := models.DB.GithubRepoRemoved(installationId, appId, repo.FullName)
		if err != nil {
			return err
		}

		// todo: change the status of DiggerRepo to InActive
	}
	return nil
}

func handleInstallationCreatedEvent(installation *webhooks.InstallationPayload) error {
	installationId := installation.Installation.ID
	login := installation.Installation.Account.Login
	accountId := installation.Installation.Account.ID
	appId := installation.Installation.AppID

	for _, repo := range installation.Repositories {
		log.Printf("Adding a new installation %d for repo: %s", installationId, repo.FullName)
		err := models.DB.GithubRepoAdded(installationId, appId, login, accountId, repo.FullName)
		if err != nil {
			return err
		}
		err = createDiggerRepoForGithubRepo(repo.FullName, installationId)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleInstallationDeletedEvent(installation *webhooks.InstallationPayload) error {
	installationId := installation.Installation.ID
	appId := installation.Installation.AppID
	for _, repo := range installation.Repositories {
		log.Printf("Removing an installation %d for repo: %s", installationId, repo.FullName)
		err := models.DB.GithubRepoRemoved(installationId, appId, repo.FullName)
		if err != nil {
			return err
		}
	}
	return nil
}

func handlePullRequestEvent(gh utils.GithubClientProvider, payload *webhooks.PullRequestPayload) error {
	var installationId int64
	var repoName string
	var repoOwner string
	var repoFullName string
	var cloneURL string

	installationId = payload.Installation.ID
	repoName = payload.Repository.Name
	repoOwner = payload.Repository.Owner.Login
	repoFullName = payload.Repository.FullName
	cloneURL = payload.Repository.CloneURL

	ghService, config, projectsGraph, branch, err := getDiggerConfig(gh, installationId, repoFullName, repoOwner, repoName, cloneURL, int(payload.PullRequest.Number))

	impactedProjects, _, err := dg_github.ProcessGitHubPullRequestEvent(payload, config, projectsGraph, ghService)
	if err != nil {
		log.Printf("Error processing event: %v", err)
		return fmt.Errorf("error processing event")
	}

	jobsForImpactedProjects, _, err := dg_github.ConvertGithubPullRequestEventToJobs(payload, impactedProjects, nil, config.Workflows)
	if err != nil {
		log.Printf("Error converting event to jobsForImpactedProjects: %v", err)
		return fmt.Errorf("error converting event to jobsForImpactedProjects")
	}

	impactedProjectsMap := make(map[string]dg_configuration.Project)
	for _, p := range impactedProjects {
		impactedProjectsMap[p.Name] = p
	}

	impactedJobsMap := make(map[string]orchestrator.Job)
	for _, j := range jobsForImpactedProjects {
		impactedJobsMap[j.ProjectName] = j
	}

	batchId, _, err := utils.ConvertJobsToDiggerJobs(impactedJobsMap, impactedProjectsMap, projectsGraph, *branch, repoFullName)
	if err != nil {
		log.Printf("ConvertJobsToDiggerJobs error: %v", err)
		return fmt.Errorf("error convertingjobs")
	}

	err = TriggerDiggerJobs(ghService.Client, repoOwner, repoName, *batchId)
	if err != nil {
		log.Printf("TriggerDiggerJobs error: %v", err)
		return fmt.Errorf("error triggerring GitHub Actions for Digger Jobs")
	}

	return nil
}

func getDiggerConfig(gh utils.GithubClientProvider, installationId int64, repoFullName string, repoOwner string, repoName string, cloneUrl string, prNumber int) (*dg_github.GithubService, *dg_configuration.DiggerConfig, graph.Graph[string, dg_configuration.Project], *string, error) {
	installation, err := models.DB.GetGithubAppInstallationByIdAndRepo(installationId, repoFullName)
	if err != nil {
		log.Printf("Error getting installation: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting installation")
	}

	_, err = models.DB.GetGithubApp(installation.GithubAppId)
	if err != nil {
		log.Printf("Error getting app: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting app")
	}

	ghClient, token, err := gh.Get(installation.GithubAppId, installation.GithubInstallationId)
	if err != nil {
		log.Printf("Error creating github app client: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error creating github app client")
	}

	ghService := dg_github.GithubService{
		Client:   ghClient,
		RepoName: repoName,
		Owner:    repoOwner,
	}

	var prBranch string
	prBranch, err = ghService.GetBranchName(prNumber)
	if err != nil {
		log.Printf("Error getting branch name: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting branch name")
	}

	link, err := models.DB.GetGithubAppInstallationLink(installationId)
	if err != nil {
		log.Printf("Error getting branch name: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting branch name")
	}

	diggerRepoName := repoOwner + "-" + repoName
	repo, err := models.DB.GetRepo(link.Organisation.ID, diggerRepoName)
	if err != nil {
		log.Printf("Error getting repo: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting repo")
	}

	configYaml, err := dg_configuration.LoadDiggerConfigYamlFromString(repo.DiggerConfig)
	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error loading digger config")
	}

	log.Printf("Digger config loadded successfully\n")

	if configYaml.GenerateProjectsConfig != nil {
		err = utils.CloneGitRepoAndDoAction(cloneUrl, prBranch, *token, func(dir string) {
			dg_configuration.HandleYamlProjectGeneration(configYaml, dir)
		})
		if err != nil {
			log.Printf("Error generating projects: %v", err)
			return nil, nil, nil, nil, fmt.Errorf("error generating projects")
		}
	}

	config, dependencyGraph, err := loadDiggerConfig(configYaml)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error loading digger config")
	}
	log.Printf("Digger config parsed successfully\n")
	return &ghService, config, dependencyGraph, &prBranch, nil
}

func handleIssueCommentEvent(gh utils.GithubClientProvider, payload *webhooks.IssueCommentPayload) error {
	var installationId int64
	var repoName string
	var repoOwner string
	var repoFullName string
	var cloneURL string

	installationId = payload.Installation.ID
	repoName = payload.Repository.Name
	repoOwner = payload.Repository.Owner.Login
	repoFullName = payload.Repository.FullName
	cloneURL = payload.Repository.CloneURL

	ghService, config, projectsGraph, branch, err := getDiggerConfig(gh, installationId, repoFullName, repoOwner, repoName, cloneURL, int(payload.Issue.Number))

	impactedProjects, requestedProject, _, err := dg_github.ProcessGitHubIssueCommentEvent(payload, config, projectsGraph, ghService)
	if err != nil {
		log.Printf("Error processing event: %v", err)
		return fmt.Errorf("error processing event")
	}
	log.Printf("GitHub IssueComment event processed successfully\n")

	jobs, _, err := dg_github.ConvertGithubIssueCommentEventToJobs(payload, impactedProjects, requestedProject, config.Workflows)
	if err != nil {
		log.Printf("Error converting event to jobs: %v", err)
		return fmt.Errorf("error converting event to jobs")
	}
	log.Printf("GitHub IssueComment event converted to Jobs successfully\n")

	impactedProjectsMap := make(map[string]dg_configuration.Project)
	for _, p := range impactedProjects {
		impactedProjectsMap[p.Name] = p
	}

	impactedProjectsJobMap := make(map[string]orchestrator.Job)
	for _, p := range impactedProjects {
		for _, j := range jobs {
			if j.ProjectName == p.Name {
				impactedProjectsJobMap[p.Name] = j
			}
		}
	}

	batchId, _, err := utils.ConvertJobsToDiggerJobs(impactedProjectsJobMap, impactedProjectsMap, projectsGraph, *branch, repoFullName)
	if err != nil {
		log.Printf("ConvertJobsToDiggerJobs error: %v", err)
		return fmt.Errorf("error convertingjobs")
	}

	err = TriggerDiggerJobs(ghService.Client, repoOwner, repoName, *batchId)
	if err != nil {
		log.Printf("TriggerDiggerJobs error: %v", err)
		return fmt.Errorf("error triggerring GitHub Actions for Digger Jobs")
	}
	return nil
}

func TriggerDiggerJobs(client *github.Client, repoOwner string, repoName string, batchId uuid.UUID) error {
	diggerJobs, err := models.DB.GetDiggerJobsWithoutParentForBatch(batchId)

	log.Printf("number of diggerJobs:%v\n", len(diggerJobs))

	for _, job := range diggerJobs {
		if job.SerializedJob == nil {
			return fmt.Errorf("GitHub job can't be nil")
		}
		jobString := string(job.SerializedJob)
		log.Printf("jobString: %v \n", jobString)
		// TODO: make workflow file name configurable
		_, err = client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), repoOwner, repoName, "workflow.yml", github.CreateWorkflowDispatchEventRequest{
			Ref:    job.BranchName,
			Inputs: map[string]interface{}{"job": jobString, "id": job.DiggerJobId},
		})
		if err != nil {
			log.Printf("failed to trigger github workflow, %v\n", err)
			return fmt.Errorf("failed to trigger github workflow, %v\n", err)
		} else {
			job.Status = models.DiggerJobTriggered
			err := models.DB.UpdateDiggerJob(&job)
			if err != nil {
				log.Printf("failed to trigger github workflow, %v\n", err)
				return fmt.Errorf("failed to trigger github workflow, %v\n", err)
			}
		}
	}
	return nil
}

func CreateDiggerWorkflow(client *github.Client, githubRepo string) error {
	ctx := context.Background()
	repoOwner := strings.Split(githubRepo, "/")[0]
	repoName := strings.Split(githubRepo, "/")[1]
	// create branch 'digger/configure'
	// Create remote branch from master

	repo, _, _ := client.Repositories.Get(ctx, repoOwner, repoName)
	defaultBranch := *repo.DefaultBranch

	defaultBranchRef, _, _ := client.Git.GetRef(ctx, repoOwner, repoName, "refs/heads/"+defaultBranch) // or "refs/heads/main"
	branch := "digger/configure"
	refName := fmt.Sprintf("refs/heads/%s", branch)
	branchRef := &github.Reference{
		Ref: &refName,
		Object: &github.GitObject{
			SHA: defaultBranchRef.Object.SHA,
		},
	}
	// trying to create a new branch
	_, _, err := client.Git.CreateRef(ctx, repoOwner, repoName, branchRef)
	if err != nil {
		// if branch already exist, do nothing
		if strings.Contains(err.Error(), "Reference already exists") {
			return nil
		}
		return fmt.Errorf("failed to create a branch, %w", err)
	}

	workflowFileContents := `on:
  workflow_dispatch:
    inputs:
      id:
        description: 'run identifier'
        required: false
      job:
        required: true
jobs:
  build:
    name: Workflow ID Provider
    runs-on: ubuntu-latest
    steps:
      - name: digger run
        uses: diggerhq/digger@develop
        with:
          setup-aws: false
          disable-locking: true
          digger-token: ${{ secrets.DIGGER_TOKEN }}
          digger-hostname: 'https://cloud.uselemon.cloud'
          digger-organisation: 'digger'
        env:
          GITHUB_CONTEXT: ${{ toJson(github) }}
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`

	// Push file
	msg := "add Digger GitHub workflow file"
	var req github.RepositoryContentFileOptions
	req.Content = []byte(workflowFileContents)
	req.Message = &msg
	req.Branch = &branch
	filePath := ".github/workflows/workflow.yml"

	opts := &github.RepositoryContentGetOptions{Ref: *defaultBranchRef.Ref}
	contents, _, _, err := client.Repositories.GetContents(ctx, repoOwner, repoName, filePath, opts)
	if err != nil {
		return err
	}
	// workflow file already exist, do nothing
	if err == nil {
		return nil
	}

	if *contents.Content != workflowFileContents {
		log.Printf("workflow file has been modified")
	}

	_, _, err = client.Repositories.CreateFile(ctx, repoOwner, repoName, filePath, &req)
	if err != nil {
		return fmt.Errorf("failed to create digger workflow file, %w", err)
	}

	prTitle := "Configure Digger"
	pullRequest := &github.NewPullRequest{Title: &prTitle,
		Head: &branch, Base: &defaultBranch}
	_, _, err = client.PullRequests.Create(ctx, repoOwner, repoName, pullRequest)
	if err != nil {
		return fmt.Errorf("failed to create a pull request for digger/configure, %w", err)
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
		log.Printf("err: %v", err)
		c.String(http.StatusInternalServerError, "Failed to parse installation_id.")
		return
	}

	result, err := validateGithubCallback(clientId, clientSecret, code, installationId64)
	if !result {
		log.Printf("Failed to validated installation id, %v\n", err)
		c.String(http.StatusInternalServerError, "Failed to validate installation_id.")
		return
	}

	org, err := models.DB.GetOrganisationById(orgId)
	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	_, err = models.DB.CreateGithubInstallationLink(org, installationId64)
	if err != nil {
		log.Printf("Error saving CreateGithubInstallationLink to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating GitHub installation"})
		return
	}
	c.Redirect(http.StatusFound, "/repos")
}

func GithubReposPage(c *gin.Context) {
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)
	if !exists {
		log.Printf("Organisation ID not found in context")
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	link, err := models.DB.GetGithubInstallationLinkForOrg(orgId)
	if err != nil {
		log.Printf("GetGithubInstallationLinkForOrg error: %v\n", err)
		c.String(http.StatusForbidden, "Failed to find any GitHub installations for this org")
		return
	}

	installations, err := models.DB.GetGithubAppInstallations(link.GithubInstallationId)
	if err != nil {
		log.Printf("GetGithubAppInstallations error: %v\n", err)
		c.String(http.StatusForbidden, "Failed to find any GitHub installations for this org")
		return
	}

	if len(installations) == 0 {
		c.String(http.StatusForbidden, "Failed to find any GitHub installations for this org")
		return
	}

	gh := &utils.DiggerGithubRealClientProvider{}
	client, _, err := gh.Get(installations[0].GithubAppId, installations[0].GithubInstallationId)
	if err != nil {
		log.Printf("GetGithubAppInstallations error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating GitHub client"})
		return
	}

	opts := &github.ListOptions{}
	repos, _, err := client.Apps.ListRepos(context.Background(), opts)
	if err != nil {
		log.Printf("GetGithubAppInstallations error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list GitHub repos."})
		return
	}
	c.HTML(http.StatusOK, "github_repos.tmpl", gin.H{"Repos": repos.Repositories})
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
		log.Printf("installation id: %v\n", *v.ID)
		if *v.ID == installationId {
			installationIdMatch = true
		}
	}
	if !installationIdMatch {
		return false, fmt.Errorf("InstallationId %v doesn't match any id for specified user\n", installationId)
	}
	return true, nil
}
