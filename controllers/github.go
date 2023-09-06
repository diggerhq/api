package controllers

import (
	"context"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bradleyfalzon/ghinstallation/v2"
	webhooks "github.com/diggerhq/webhooks/github"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v54/github"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func GitHubAppWebHook(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	hook, _ := webhooks.New()

	payload, err := hook.Parse(c.Request, webhooks.InstallationEvent, webhooks.PullRequestEvent, webhooks.IssueCommentEvent,
		webhooks.InstallationRepositoriesEvent, webhooks.WorkflowJobEvent, webhooks.WorkflowRunEvent)
	if err != nil {
		if errors.Is(err, webhooks.ErrEventNotFound) {
			// ok event wasn't one of the ones asked to be parsed
			fmt.Println("GitHub event  wasn't found.")
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
				c.String(http.StatusInternalServerError, "Failed to store item.")
				return
			}
		}

		if installation.Action == "deleted" {
			err := handleInstallationDeletedEvent(installation)
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to remove item.")
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
				err := models.GitHubRepoAdded(installationId, appId, login, accountId, repo.FullName)
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
				err := models.GitHubRepoRemoved(installationId, appId, repo.FullName)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to remove item.")
					return
				}
			}
		}

	case webhooks.IssueCommentPayload:
		issueComment := payload.(webhooks.IssueCommentPayload)
		fmt.Printf("new comment: %+v", issueComment)
	case webhooks.WorkflowJobPayload:
		payload := payload.(webhooks.WorkflowJobPayload)
		err := handleWorkflowJobEvent(payload)
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
	}

	c.JSON(200, "ok")
}

func getGitHubClient(githubAppId int64, installationId int64) (*github.Client, error) {
	githubAppPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	client, err := GetGithubClient(githubAppId, installationId, githubAppPrivateKey)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func handleInstallationCreatedEvent(installation webhooks.InstallationPayload) error {
	installationId := installation.Installation.ID
	login := installation.Installation.Account.Login
	accountId := installation.Installation.Account.ID
	appId := installation.Installation.AppID

	for _, repo := range installation.Repositories {
		fmt.Printf("Adding a new installation %d for repo: %s", installationId, repo.FullName)
		err := models.GitHubRepoAdded(installationId, appId, login, accountId, repo.FullName)
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
		err := models.GitHubRepoRemoved(installationId, appId, repo.FullName)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleWorkflowJobEvent(payload webhooks.WorkflowJobPayload) error {
	ctx := context.Background()
	switch payload.Action {
	case "completed":
		githubJobId := payload.WorkflowJob.ID
		//githubJobStatus := payload.WorkflowJob.Status

		repo := payload.Repository.Name
		owner := payload.Repository.Owner.Login
		repoFullName := payload.Repository.FullName
		installationId := payload.Installation.ID

		installation, err := models.GetGitHubAppInstallationByIdAndRepo(installationId, repoFullName)
		if err != nil {
			return err
		}
		client, err := getGitHubClient(installation.GithubAppId, installationId)
		if err != nil {
			return err
		}

		workflowJob, _, err := client.Actions.GetWorkflowJobByID(ctx, owner, repo, githubJobId)
		if err != nil {
			return err
		}

		for _, s := range (*workflowJob).Steps {
			name := *s.Name
			if strings.HasPrefix(name, "digger run ") {
				// digger job id and workflow step name matched
				jobId := strings.Replace(name, "digger run ", "", 1)
				_, err := models.UpdateDiggerJob(repoFullName, jobId, githubJobId)
				if err != nil {
					return err
				}
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

func GitHubAppCallbackPage(c *gin.Context) {
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

	result, err := validateGitHubCallback(clientId, clientSecret, code, installationId64)
	if !result {
		fmt.Printf("Failed to validated installation id, %v\n", err)
		c.String(http.StatusInternalServerError, "Failed to validate installation_id.")
		return
	}

	org, err := models.GetOrganisationById(orgId)
	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	_, err = models.CreateGitHubInstallationLink(org.ID, installationId64)
	if err != nil {
		log.Printf("Error saving CreateGitHubInstallationLink to database: %v", err)
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

	owner := "diggerhq"
	repo := "github-job-scheduler"
	workflowFileName := "plan.yml"
	repoFullName := owner + "/" + repo

	installation, err := models.GetGitHubAppInstallationByOrgAndRepo(orgId, repoFullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating github installation"})
		return
	}

	githubAppPrivateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	link, err := models.CreateDiggerJob(repoFullName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating a test job"})
		return
	}

	client, err := GetGithubClient(installation.GithubAppId, installation.GithubInstallationId, githubAppPrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating a token"})
		return
	}

	TriggerTestJob(client, owner, repo, link.DiggerJobId, workflowFileName)
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
}

func GetGithubClient(githubAppId int64, installationId int64, githubAppPrivateKey string) (*github.Client, error) {
	tr := http.DefaultTransport
	itr, err := ghinstallation.New(tr, githubAppId, installationId, []byte(githubAppPrivateKey))
	if err != nil {
		return nil, fmt.Errorf("error initialising installation: %v\n", err)
	}

	ghClient := github.NewClient(&http.Client{Transport: itr})
	return ghClient, nil
}

func TriggerTestJob(client *github.Client, repoOwner string, repoName string, jobId string, workflowFileName string) {
	//_, _, _ := client.Repositories.Get(ctx, owner, repo_name)
	ctx := context.Background()
	event := github.CreateWorkflowDispatchEventRequest{Ref: "main", Inputs: map[string]interface{}{"id": jobId}}
	_, err := client.Actions.CreateWorkflowDispatchEventByFileName(ctx, repoOwner, repoName, workflowFileName, event)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
}

// why this validation is needed: https://roadie.io/blog/avoid-leaking-github-org-data/
// validation based on https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-user-access-token-for-a-github-app , step 3
func validateGitHubCallback(clientId string, clientSecret string, code string, installationId int64) (bool, error) {
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
