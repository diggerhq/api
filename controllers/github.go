package controllers

import (
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"io"
	"net/http"
)

func GitHubAppCallback(c *gin.Context) {
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}

	c.Header("Content-Type", "application/json")
	fmt.Print("---------- github app callback ---------------- ")
	fmt.Printf(string(requestBody))
	c.JSON(200, "ok")
}

func GitHubAppWebHook(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	hook, _ := github.New()

	payload, err := hook.Parse(c.Request, github.InstallationEvent, github.PullRequestEvent, github.IssueCommentEvent, github.InstallationRepositoriesEvent)
	if err != nil {
		if errors.Is(err, github.ErrEventNotFound) {
			// ok event wasn't one of the ones asked to be parsed
			fmt.Println("GitHub event  wasn't found.")
		}
		fmt.Printf("Failed to parse Github Event. :%v", err)
		c.String(http.StatusInternalServerError, "Failed to parse Github Event")
		return
	}
	switch payload.(type) {

	case github.InstallationPayload:
		fmt.Println("case github.InstallationPayload:")
		installation := payload.(github.InstallationPayload)
		if installation.Action == "created" {
			installationId := installation.Installation.ID
			login := installation.Installation.Account.Login
			accountId := installation.Installation.Account.ID
			appId := installation.Installation.AppID
			fmt.Printf("accountId: %d\n", accountId)

			for _, repo := range installation.Repositories {
				err := models.GitHubRepoAdded(installationId, appId, login, accountId, repo.FullName)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to store item.")
					return
				}
			}
		}

		if installation.Action == "deleted" {
			installationId := installation.Installation.ID
			appId := installation.Installation.AppID
			for _, repo := range installation.Repositories {
				err := models.GitHubRepoRemoved(installationId, appId, repo.FullName)
				if err != nil {
					c.String(http.StatusInternalServerError, "Failed to remove item.")
					return
				}
			}
		}
	case github.InstallationRepositoriesPayload:
		installationRepos := payload.(github.InstallationRepositoriesPayload)
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

	case github.IssueCommentPayload:
		issueComment := payload.(github.IssueCommentPayload)
		// Do whatever you want from here...
		fmt.Printf("new comment: %+v", issueComment)
	}

	c.JSON(200, "ok")
}

func GitHubAppSetupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
}
