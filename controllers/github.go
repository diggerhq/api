package controllers

import (
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/webhooks/v6/github"
	"gorm.io/gorm"
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
				err := repoAdded(installationId, appId, login, accountId, repo.FullName)
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
				err := repoRemoved(installationId, appId, repo.FullName)
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
				err := repoAdded(installationId, appId, login, accountId, repo.FullName)
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
				err := repoRemoved(installationId, appId, repo.FullName)
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

func repoAdded(installationId int64, appId int, login string, accountId int64, repoFullName string) error {
	// check if item exist already
	item := models.GithubAppInstallation{}
	result := models.DB.Where("github_installation_id = ? AND repo=?", installationId, repoFullName).First(&item)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to find github installation in database. %v", result.Error)
		}
	}

	if result.RowsAffected == 0 {
		item := models.GithubAppInstallation{
			GithubInstallationId: installationId,
			GithubAppId:          int64(appId),
			Login:                login,
			AccountId:            int(accountId),
			Repo:                 repoFullName,
			State:                models.Active,
		}
		err := models.DB.Create(&item).Error
		if err != nil {
			fmt.Printf("Failed to save github installation item to database. %v\n", err)
			return fmt.Errorf("failed to save github installation item to database. %v", err)
		}
	} else {
		fmt.Printf("Record for installation_id: %d, repo: %s, with state=active exist already.", installationId, repoFullName)
		item.State = models.Active
		err := models.DB.Save(item).Error
		if err != nil {
			return fmt.Errorf("failed to update github installation in the database. %v", err)
		}
	}
	return nil
}

func repoRemoved(installationId int64, appId int, repoFullName string) error {
	item := models.GithubAppInstallation{}
	err := models.DB.Where("github_installation_id = ? AND state=? AND github_app_id=? AND repo=?", installationId, models.Active, appId, repoFullName).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Record not found for installationId: %d, state=active, githubAppId: %d and repo: %s", installationId, appId, repoFullName)
			return nil
		}
		return fmt.Errorf("failed to find github installation in database. %v", err)
	}
	item.State = models.Deleted
	err = models.DB.Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update github installation in the database. %v", err)
	}
	return nil
}

func GitHubAppSetupPage(c *gin.Context) {
	c.HTML(http.StatusOK, "github_setup.tmpl", gin.H{})
}
