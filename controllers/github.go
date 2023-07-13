package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"time"
)

type GitHubInstallationNotification struct {
	Action       string `json:"action"`
	Installation struct {
		Id      int `json:"id"`
		Account struct {
			Login             string `json:"login"`
			Id                int    `json:"id"`
			NodeId            string `json:"node_id"`
			AvatarUrl         string `json:"avatar_url"`
			GravatarId        string `json:"gravatar_id"`
			Url               string `json:"url"`
			HtmlUrl           string `json:"html_url"`
			FollowersUrl      string `json:"followers_url"`
			FollowingUrl      string `json:"following_url"`
			GistsUrl          string `json:"gists_url"`
			StarredUrl        string `json:"starred_url"`
			SubscriptionsUrl  string `json:"subscriptions_url"`
			OrganizationsUrl  string `json:"organizations_url"`
			ReposUrl          string `json:"repos_url"`
			EventsUrl         string `json:"events_url"`
			ReceivedEventsUrl string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"account"`
		RepositorySelection string `json:"repository_selection"`
		AccessTokensUrl     string `json:"access_tokens_url"`
		RepositoriesUrl     string `json:"repositories_url"`
		HtmlUrl             string `json:"html_url"`
		AppId               int    `json:"app_id"`
		AppSlug             string `json:"app_slug"`
		TargetId            int    `json:"target_id"`
		TargetType          string `json:"target_type"`
		Permissions         struct {
			Issues           string `json:"issues"`
			Actions          string `json:"actions"`
			Secrets          string `json:"secrets"`
			Metadata         string `json:"metadata"`
			Statuses         string `json:"statuses"`
			Workflows        string `json:"workflows"`
			PullRequests     string `json:"pull_requests"`
			RepositoryHooks  string `json:"repository_hooks"`
			ActionsVariables string `json:"actions_variables"`
		} `json:"permissions"`
		Events                 []string      `json:"events"`
		CreatedAt              time.Time     `json:"created_at"`
		UpdatedAt              time.Time     `json:"updated_at"`
		HasMultipleSingleFiles bool          `json:"has_multiple_single_files"`
		SingleFilePaths        []interface{} `json:"single_file_paths"`
	} `json:"installation"`
	Repositories []struct {
		Id       int    `json:"id"`
		NodeId   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	} `json:"repositories"`
	Sender struct {
		Login             string `json:"login"`
		Id                int    `json:"id"`
		NodeId            string `json:"node_id"`
		AvatarUrl         string `json:"avatar_url"`
		GravatarId        string `json:"gravatar_id"`
		Url               string `json:"url"`
		HtmlUrl           string `json:"html_url"`
		FollowersUrl      string `json:"followers_url"`
		FollowingUrl      string `json:"following_url"`
		GistsUrl          string `json:"gists_url"`
		StarredUrl        string `json:"starred_url"`
		SubscriptionsUrl  string `json:"subscriptions_url"`
		OrganizationsUrl  string `json:"organizations_url"`
		ReposUrl          string `json:"repos_url"`
		EventsUrl         string `json:"events_url"`
		ReceivedEventsUrl string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"sender"`
}

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
			c.String(http.StatusInternalServerError, "Error reading request body")
			return
		}

		notification := GitHubInstallationNotification{}
		err = json.Unmarshal(requestBody, &notification)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to parse request's JSON")
			return
		}

		c.Header("Content-Type", "application/json")
		fmt.Print("---------- github app webhook ---------------- ")
		fmt.Printf("notification: %v", notification)
		c.JSON(200, "ok")
	}
}
