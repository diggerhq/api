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
	"github.com/dominikbraun/graph"
	"github.com/google/uuid"

	dg_configuration "github.com/diggerhq/lib-digger-config"
	orchestrator "github.com/diggerhq/lib-orchestrator"
	dg_github "github.com/diggerhq/lib-orchestrator/github"
	webhooks "github.com/diggerhq/webhooks/github"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GithubAppWebHook(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	gh := &utils.DiggerGithubRealClient{}

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
	case webhooks.WorkflowJobPayload:
		payload := payload.(webhooks.WorkflowJobPayload)
		err := handleWorkflowJobEvent(gh, &payload)
		if err != nil {
			log.Printf("handleWorkflowJobEvent error: %v", err)
			c.String(http.StatusInternalServerError, "Failed to handle WorkflowJob event.")
			return
		}
	case webhooks.WorkflowRunPayload:
		payload := payload.(webhooks.WorkflowRunPayload)
		err := handleWorkflowRunEvent(payload)
		if err != nil {
			log.Printf("handleWorkflowRunEvent error: %v", err)
			c.String(http.StatusInternalServerError, "Failed to handle WorkflowRun event.")
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

func handleWorkflowJobEvent(gh utils.DiggerGithubClient, payload *webhooks.WorkflowJobPayload) error {

	log.Printf("handleWorkflowJobEvent\n")
	ctx := context.Background()
	switch payload.Action {
	case "completed":
		log.Printf("handleWorkflowJobEvent completed\n")
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

		log.Printf("repoFullName: %v\n", repoFullName)
		var jobId string
		for _, s := range (*workflowJob).Steps {

			name := *s.Name
			log.Printf("workflow step: %v\n", name)
			if strings.HasPrefix(name, "digger run ") {

				// digger job id and workflow step name matched
				jobId = strings.Replace(name, "digger run ", "", 1)
				log.Printf("workflow step match, jobId %v\n", jobId)
				_, err := models.DB.UpdateDiggerJobLink(jobId, repoFullName, githubJobId)
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

	installationId = payload.Installation.ID
	repoName = payload.Repository.Name
	repoOwner = payload.Repository.Owner.Login
	repoFullName = payload.Repository.FullName
	cloneURL = payload.Repository.CloneURL

	ghService, config, projectsGraph, branch, err := getDiggerConfig(gh, installationId, repoFullName, repoOwner, repoName, cloneURL, int(payload.PullRequest.Number))

	impactedProjects, requestedProject, _, err := dg_github.ProcessGitHubPullRequestEvent(payload, config, ghService)
	if err != nil {
		log.Printf("Error processing event: %v", err)
		return fmt.Errorf("error processing event")
	}

	jobs, _, err := dg_github.ConvertGithubPullRequestEventToJobs(payload, impactedProjects, requestedProject, config.Workflows)
	if err != nil {
		log.Printf("Error converting event to jobs: %v", err)
		return fmt.Errorf("error converting event to jobs")
	}

	jobsMap := make(map[string]orchestrator.Job)
	for _, p := range impactedProjects {
		for _, j := range jobs {
			if j.ProjectName == p.Name {
				jobsMap[p.Name] = j
			}
		}
	}

	_, err = ConvertJobsToDiggerJobs(jobsMap, projectsGraph, *branch, repoFullName)
	if err != nil {
		log.Printf("ConvertJobsToDiggerJobs error: %v", err)
		return fmt.Errorf("error convertingjobs")
	}

	err = TriggerDiggerJobs(ghService.Client, repoOwner, repoName)
	if err != nil {
		log.Printf("TriggerDiggerJobs error: %v", err)
		return fmt.Errorf("error triggerring GitHub Actions for Digger Jobs")
	}

	return nil
}

func getDiggerConfig(gh utils.DiggerGithubClient, installationId int64, repoFullName string, repoOwner string, repoName string, cloneUrl string, prNumber int) (*dg_github.GithubService, *dg_configuration.DiggerConfig, graph.Graph[string, string], *string, error) {
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

	ghClient, token, err := gh.GetGithubClient(installation.GithubAppId, installation.GithubInstallationId)
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

	config, graph, err := loadDiggerConfig(configYaml)

	if err != nil {
		log.Printf("Error loading digger config: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error loading digger config")
	}
	log.Printf("Digger config parsed successfully\n")
	return &ghService, config, graph, &prBranch, nil
}

func handleIssueCommentEvent(gh utils.DiggerGithubClient, payload *webhooks.IssueCommentPayload) error {
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

	impactedProjects, requestedProject, _, err := dg_github.ProcessGitHubIssueCommentEvent(payload, config, ghService)
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

	jobsMap := make(map[string]orchestrator.Job)
	for _, p := range impactedProjects {
		for _, j := range jobs {
			if j.ProjectName == p.Name {
				jobsMap[p.Name] = j
			}
		}
	}

	_, err = ConvertJobsToDiggerJobs(jobsMap, projectsGraph, *branch, repoFullName)
	if err != nil {
		log.Printf("ConvertJobsToDiggerJobs error: %v", err)
		return fmt.Errorf("error convertingjobs")
	}

	err = TriggerDiggerJobs(ghService.Client, repoOwner, repoName)
	if err != nil {
		log.Printf("TriggerDiggerJobs error: %v", err)
		return fmt.Errorf("error triggerring GitHub Actions for Digger Jobs")
	}
	return nil
}

// ConvertJobsToDiggerJobs jobs is map with project name as a key and a Job as a value
func ConvertJobsToDiggerJobs(jobsMap map[string]orchestrator.Job, projectsGraph graph.Graph[string, string], branch string, repoFullName string) (map[string]*models.DiggerJob, error) {
	result := make(map[string]*models.DiggerJob)

	log.Printf("Number of Jobs: %v\n", len(jobsMap))
	marshalledJobsMap := map[string][]byte{}
	for _, job := range jobsMap {
		marshalled, _ := json.Marshal(orchestrator.JobToJson(job))
		marshalledJobsMap[job.ProjectName] = marshalled
	}

	batchId, _ := uuid.NewUUID()
	predecessorMap, _ := projectsGraph.PredecessorMap()

	visit := func(value string) bool {
		// value is project name

		// does it have a parent?
		if predecessorMap[value] == nil || len(predecessorMap[value]) == 0 {
			fmt.Printf("no parent for %v\n", value)
			if result[value] == nil {
				fmt.Printf("no diggerjob has been created for %v\n", value)
				// we found a node without parent, we can create a digger job
				parentJob, err := models.DB.CreateDiggerJob(batchId, nil, marshalledJobsMap[value], branch)
				if err != nil {
					log.Printf("failed to create a job")
					return false
				}
				_, err = models.DB.CreateDiggerJobLink(parentJob.DiggerJobId, repoFullName)
				if err != nil {
					log.Printf("failed to create a digger job link")
					return false
				}
				result[value] = parentJob
			}
		} else {
			// we found a node with parent(s), parent should be in results already
			parents := predecessorMap[value]
			for _, edge := range parents {
				parent := edge.Source
				fmt.Printf("parent: %v\n", parent)
				parentDiggerJob := result[parent]
				childJob, err := models.DB.CreateDiggerJob(batchId, &parentDiggerJob.DiggerJobId, marshalledJobsMap[value], branch)
				if err != nil {
					log.Printf("failed to create a job")
					return false
				}
				_, err = models.DB.CreateDiggerJobLink(childJob.DiggerJobId, repoFullName)
				if err != nil {
					log.Printf("failed to create a digger job link")
					return false
				}
				result[value] = childJob
			}
		}
		return false
	}

	log.Printf("len of predecessorMap: %v\n", len(predecessorMap))

	for node := range predecessorMap {
		if predecessorMap[node] == nil || len(predecessorMap[node]) == 0 {
			err := graph.DFS(projectsGraph, node, visit)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func TriggerDiggerJobs(client *github.Client, repoOwner string, repoName string) error {
	diggerJobs, err := models.DB.GetDiggerJobsWithoutParent()

	log.Printf("number of diggerJobs:%v\n", len(diggerJobs))

	for _, job := range diggerJobs {
		if job.SerializedJob == nil {
			return fmt.Errorf("GitHub job can't me nil")
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
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
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

	gh := &utils.DiggerGithubRealClient{}
	client, _, err := gh.GetGithubClient(installations[0].GithubAppId, installations[0].GithubInstallationId)
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
