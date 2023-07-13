package controllers

import (
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"io"
	"net/http"
)

func GitHubAppCallback() func(c *gin.Context) {
	return func(c *gin.Context) {
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
}

func GitHubAppWebHook() func(c *gin.Context) {
	return func(c *gin.Context) {
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Printf("Error reading request body. %v\n", err)
			c.String(http.StatusInternalServerError, "Error reading request body")
			return
		}

		fmt.Printf("webhook request: %s", string(requestBody))

		hook, _ := github.New()

		payload, err := hook.Parse(c.Request, github.InstallationEvent, github.PullRequestEvent, github.IssueCommentEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				// ok event wasn't one of the ones asked to be parsed
			}
		}
		switch payload.(type) {

		case github.InstallationPayload:
			installation := payload.(github.InstallationPayload)
			if installation.Action == "created" {
				installationId := installation.Installation.ID
				login := installation.Installation.Account.Login
				accountId := installation.Installation.Account.ID

				for _, repo := range installation.Repositories {
					item := models.GitHubAppInstallation{
						InstallationId: int(installationId),
						Login:          login,
						AccountId:      int(accountId),
						Repo:           repo.FullName,
						State:          models.Active,
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
				login := installation.Installation.Account.Login
				accountId := installation.Installation.Account.ID

				for _, repo := range installation.Repositories {
					item := models.GitHubAppInstallation{
						InstallationId: int(installationId),
						Login:          login,
						AccountId:      int(accountId),
						Repo:           repo.FullName,
						State:          models.Deleted,
					}
					err := models.DB.Create(&item).Error
					if err != nil {
						fmt.Printf("Failed to save record to database. %v\n", err)
						c.String(http.StatusInternalServerError, "Failed to save record to database.")
						return
					}
				}
			}
		case github.IssueCommentPayload:
			issueComment := payload.(github.IssueCommentPayload)
			// Do whatever you want from here...
			fmt.Printf("new comment: %+v", issueComment)
		}
		c.Header("Content-Type", "application/json")
		c.JSON(200, "ok")
	}
}
