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

	payload, err := hook.Parse(c.Request, github.InstallationEvent, github.PullRequestEvent, github.IssueCommentEvent)
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
			fmt.Printf("accountId: %d\n", accountId)

			for _, repo := range installation.Repositories {
				item := models.GithubAppInstallation{
					GithubInstallationId: installationId,
					Login:                login,
					AccountId:            int(accountId),
					Repo:                 repo.FullName,
					State:                models.Active,
				}
				err := models.DB.Create(&item).Error
				if err != nil {
					fmt.Printf("Failed to save record to database. %v\n", err)
					c.String(http.StatusInternalServerError, "Failed to save record to database.")
					return
				}
			}

		}

		if installation.Action == "deleted" {
			installationId := installation.Installation.ID
			accountId := installation.Installation.Account.ID
			fmt.Printf("accountId: %d\n", accountId)

			for _, repo := range installation.Repositories {
				item := models.GithubAppInstallation{}
				models.DB.Where("github_installation_id = ? AND state=? AND repo=?", installationId, models.Active, repo).First(&item)
				err := models.DB.Create(&item).Error
				if err != nil {
					fmt.Printf("Failed to find github installationin database. %v\n", err)
					c.String(http.StatusInternalServerError, "Failed to find github installation.")
					return
				}
				item.State = models.Deleted
				err = models.DB.Save(item).Error
				if err != nil {
					fmt.Printf("Failed to update github installationin in database. %v\n", err)
					c.String(http.StatusInternalServerError, "Failed to update github installation.")
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
