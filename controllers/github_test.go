package controllers

import (
	"digger.dev/cloud/models"
	"digger.dev/cloud/utils"
	"encoding/json"
	webhooks "github.com/diggerhq/webhooks/github"
	"github.com/google/go-github/v55/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"testing"
)

var issueCommentPayload string = `{
  "action": "created",
  "issue": {
    "url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2",
    "repository_url": "https://api.github.com/repos/diggerhq/github-job-scheduler",
    "labels_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2/labels{/name}",
    "comments_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2/comments",
    "events_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2/events",
    "html_url": "https://github.com/diggerhq/github-job-scheduler/pull/2",
    "id": 1882391909,
    "node_id": "33333",
    "number": 2,
    "title": "Update main.tf",
    "user": {
      "login": "veziak",
      "id": 2407061,
      "node_id": "4444=",
      "avatar_url": "https://avatars.githubusercontent.com/u/2407061?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/veziak",
      "html_url": "https://github.com/veziak",
      "followers_url": "https://api.github.com/users/veziak/followers",
      "following_url": "https://api.github.com/users/veziak/following{/other_user}",
      "gists_url": "https://api.github.com/users/veziak/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/veziak/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/veziak/subscriptions",
      "organizations_url": "https://api.github.com/users/veziak/orgs",
      "repos_url": "https://api.github.com/users/veziak/repos",
      "events_url": "https://api.github.com/users/veziak/events{/privacy}",
      "received_events_url": "https://api.github.com/users/veziak/received_events",
      "type": "User",
      "site_admin": false
    },
    "labels": [
    ],
    "state": "open",
    "locked": false,
    "assignee": null,
    "assignees": [

    ],
    "milestone": null,
    "comments": 2,
    "created_at": "2023-09-05T16:53:52Z",
    "updated_at": "2023-09-11T14:33:42Z",
    "closed_at": null,
    "author_association": "CONTRIBUTOR",
    "active_lock_reason": null,
    "draft": false,
    "pull_request": {
      "url": "https://api.github.com/repos/diggerhq/github-job-scheduler/pulls/2",
      "html_url": "https://github.com/diggerhq/github-job-scheduler/pull/2",
      "diff_url": "https://github.com/diggerhq/github-job-scheduler/pull/2.diff",
      "patch_url": "https://github.com/diggerhq/github-job-scheduler/pull/2.patch",
      "merged_at": null
    },
    "body": null,
    "reactions": {
      "url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2/reactions",
      "total_count": 0,
      "+1": 0,
      "-1": 0,
      "laugh": 0,
      "hooray": 0,
      "confused": 0,
      "heart": 0,
      "rocket": 0,
      "eyes": 0
    },
    "timeline_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2/timeline",
    "performed_via_github_app": null,
    "state_reason": null
  },
  "comment": {
    "url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/comments/1714014480",
    "html_url": "https://github.com/diggerhq/github-job-scheduler/pull/2#issuecomment-1714014480",
    "issue_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/2",
    "id": 1714014480,
    "node_id": "44444",
    "user": {
      "login": "veziak",
      "id": 2407061,
      "node_id": "33333=",
      "avatar_url": "https://avatars.githubusercontent.com/u/2407061?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/veziak",
      "html_url": "https://github.com/veziak",
      "followers_url": "https://api.github.com/users/veziak/followers",
      "following_url": "https://api.github.com/users/veziak/following{/other_user}",
      "gists_url": "https://api.github.com/users/veziak/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/veziak/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/veziak/subscriptions",
      "organizations_url": "https://api.github.com/users/veziak/orgs",
      "repos_url": "https://api.github.com/users/veziak/repos",
      "events_url": "https://api.github.com/users/veziak/events{/privacy}",
      "received_events_url": "https://api.github.com/users/veziak/received_events",
      "type": "User",
      "site_admin": false
    },
    "created_at": "2023-09-11T14:33:42Z",
    "updated_at": "2023-09-11T14:33:42Z",
    "author_association": "CONTRIBUTOR",
    "body": "digger plan",
    "reactions": {
      "url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/comments/1714014480/reactions",
      "total_count": 0,
      "+1": 0,
      "-1": 0,
      "laugh": 0,
      "hooray": 0,
      "confused": 0,
      "heart": 0,
      "rocket": 0,
      "eyes": 0
    },
    "performed_via_github_app": null
  },
  "repository": {
    "id": 686968600,
    "node_id": "222222",
    "name": "github-job-scheduler",
    "full_name": "diggerhq/github-job-scheduler",
    "private": true,
    "owner": {
      "login": "diggerhq",
      "id": 71334590,
      "node_id": "333",
      "avatar_url": "https://avatars.githubusercontent.com/u/71334590?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/diggerhq",
      "html_url": "https://github.com/diggerhq",
      "followers_url": "https://api.github.com/users/diggerhq/followers",
      "following_url": "https://api.github.com/users/diggerhq/following{/other_user}",
      "gists_url": "https://api.github.com/users/diggerhq/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/diggerhq/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/diggerhq/subscriptions",
      "organizations_url": "https://api.github.com/users/diggerhq/orgs",
      "repos_url": "https://api.github.com/users/diggerhq/repos",
      "events_url": "https://api.github.com/users/diggerhq/events{/privacy}",
      "received_events_url": "https://api.github.com/users/diggerhq/received_events",
      "type": "Organization",
      "site_admin": false
    },
    "html_url": "https://github.com/diggerhq/github-job-scheduler",
    "description": null,
    "fork": false,
    "url": "https://api.github.com/repos/diggerhq/github-job-scheduler",
    "forks_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/forks",
    "keys_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/teams",
    "hooks_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/hooks",
    "issue_events_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/events{/number}",
    "events_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/events",
    "assignees_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/assignees{/user}",
    "branches_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/branches{/branch}",
    "tags_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/tags",
    "blobs_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/languages",
    "stargazers_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/stargazers",
    "contributors_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/contributors",
    "subscribers_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/subscribers",
    "subscription_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/subscription",
    "commits_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/contents/{+path}",
    "compare_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/merges",
    "archive_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/downloads",
    "issues_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/issues{/number}",
    "pulls_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/labels{/name}",
    "releases_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/releases{/id}",
    "deployments_url": "https://api.github.com/repos/diggerhq/github-job-scheduler/deployments",
    "created_at": "2023-09-04T10:29:28Z",
    "updated_at": "2023-09-05T16:06:16Z",
    "pushed_at": "2023-09-06T17:02:35Z",
    "git_url": "git://github.com/diggerhq/github-job-scheduler.git",
    "ssh_url": "git@github.com:diggerhq/github-job-scheduler.git",
    "clone_url": "https://github.com/diggerhq/github-job-scheduler.git",
    "svn_url": "https://github.com/diggerhq/github-job-scheduler",
    "homepage": null,
    "size": 9,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": "HCL",
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": false,
    "has_discussions": false,
    "forks_count": 0,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 1,
    "license": null,
    "allow_forking": false,
    "is_template": false,
    "web_commit_signoff_required": false,
    "topics": [

    ],
    "visibility": "private",
    "forks": 0,
    "open_issues": 1,
    "watchers": 0,
    "default_branch": "main"
  },
  "organization": {
    "login": "diggerhq",
    "id": 71334590,
    "node_id": "2222",
    "url": "https://api.github.com/orgs/diggerhq",
    "repos_url": "https://api.github.com/orgs/diggerhq/repos",
    "events_url": "https://api.github.com/orgs/diggerhq/events",
    "hooks_url": "https://api.github.com/orgs/diggerhq/hooks",
    "issues_url": "https://api.github.com/orgs/diggerhq/issues",
    "members_url": "https://api.github.com/orgs/diggerhq/members{/member}",
    "public_members_url": "https://api.github.com/orgs/diggerhq/public_members{/member}",
    "avatar_url": "https://avatars.githubusercontent.com/u/71334590?v=4",
    "description": ""
  },
  "sender": {
    "login": "veziak",
    "id": 2407061,
    "node_id": "2222=",
    "avatar_url": "https://avatars.githubusercontent.com/u/2407061?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/veziak",
    "html_url": "https://github.com/veziak",
    "followers_url": "https://api.github.com/users/veziak/followers",
    "following_url": "https://api.github.com/users/veziak/following{/other_user}",
    "gists_url": "https://api.github.com/users/veziak/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/veziak/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/veziak/subscriptions",
    "organizations_url": "https://api.github.com/users/veziak/orgs",
    "repos_url": "https://api.github.com/users/veziak/repos",
    "events_url": "https://api.github.com/users/veziak/events{/privacy}",
    "received_events_url": "https://api.github.com/users/veziak/received_events",
    "type": "User",
    "site_admin": false
  },
  "installation": {
    "id": 41584295,
    "node_id": "111"
  }
}`

func setupSuite(tb testing.TB) (func(tb testing.TB), *models.Database) {
	log.Println("setup suite")

	// database file name
	dbName := "database_test.db"

	// remove old database
	e := os.Remove(dbName)
	if e != nil {
		if !strings.Contains(e.Error(), "no such file or directory") {
			log.Fatal(e)
		}
	}

	// open and create a new database
	gdb, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// migrate tables
	err = gdb.AutoMigrate(&models.Policy{}, &models.Organisation{}, &models.Repo{}, &models.Project{}, &models.Token{},
		&models.User{}, &models.ProjectRun{}, &models.GithubAppInstallation{}, &models.GithubApp{}, &models.GithubAppInstallationLink{},
		&models.GithubDiggerJobLink{}, &models.DiggerJob{})
	if err != nil {
		log.Fatal(err)
	}

	database := &models.Database{GormDB: gdb}

	orgTenantId := "11111111-1111-1111-1111-111111111111"
	externalSource := "test"
	orgName := "testOrg"
	org, err := database.CreateOrganisation(orgName, externalSource, orgTenantId)
	if err != nil {
		log.Fatal(err)
	}

	repoName := "test repo"
	repo, err := database.CreateRepo(repoName, org, "")
	if err != nil {
		log.Fatal(err)
	}

	projectName := "test project"
	_, err = database.CreateProject(projectName, org, repo)
	if err != nil {
		log.Fatal(err)
	}

	var payload webhooks.IssueCommentPayload
	err = json.Unmarshal([]byte(issueCommentPayload), &payload)
	if err != nil {
		log.Fatal(err)
	}
	// read installationID from test payload
	installationId := payload.Installation.ID

	_, err = database.CreateGithubInstallationLink(org, installationId)
	if err != nil {
		log.Fatal(err)
	}

	githubAppId := int64(1)
	login := "test"
	accountId := 1
	repoFullName := "diggerhq/github-job-scheduler"
	_, err = database.CreateGithubAppInstallation(installationId, githubAppId, login, accountId, repoFullName)
	if err != nil {
		log.Fatal(err)
	}

	diggerConfig := `projects:
- name: dev
  dir: dev
  workflow: default
- name: prod
  dir: prod
  workflow: default
  depends_on: ["dev"]
`

	diggerRepoName := strings.Replace(repoFullName, "/", "-", 1)
	_, err = database.CreateRepo(diggerRepoName, org, diggerConfig)
	if err != nil {
		log.Fatal(err)
	}

	models.DB = database
	// Return a function to teardown the test
	return func(tb testing.TB) {
		log.Println("teardown suite")
	}, database
}

func TestGithubHandleIssueCommentEvent(t *testing.T) {
	teardownSuite, _ := setupSuite(t)
	defer teardownSuite(t)

	files := make([]github.CommitFile, 2)
	files[0] = github.CommitFile{Filename: github.String("prod/main.tf")}
	files[1] = github.CommitFile{Filename: github.String("dev/main.tf")}
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			github.PullRequest{
				Number: github.Int(1),
				Head:   &github.PullRequestBranch{Ref: github.String("main")},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
			files,
		),
		mock.WithRequestMatch(
			mock.PostReposActionsWorkflowsDispatchesByOwnerByRepoByWorkflowId,
			nil,
		),
	)

	gh := &utils.DiggerGithubClientMock{}
	gh.MockedHTTPClient = mockedHTTPClient

	var payload webhooks.IssueCommentPayload
	err := json.Unmarshal([]byte(issueCommentPayload), &payload)
	assert.NoError(t, err)
	err = handleIssueCommentEvent(gh, &payload)
	assert.NoError(t, err)

	jobs, err := models.DB.GetDiggerJobsWithoutParent()
	assert.Equal(t, 1, len(jobs))
}
