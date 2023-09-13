package models

import "gorm.io/gorm"

type GithubApp struct {
	gorm.Model
	GithubId       int64
	OrganisationId int
	PrivateKey     string
}

type GithubAppInstallState int

const (
	Active  GithubAppInstallState = 1
	Deleted GithubAppInstallState = 2
)

type DiggerJobStatus int8

const (
	DiggerJobCreated   DiggerJobStatus = 1
	DiggerJobSucceeded DiggerJobStatus = 2
	DiggerJobFailed    DiggerJobStatus = 3
	DiggerJobStarted   DiggerJobStatus = 4
)

type GithubAppInstallation struct {
	gorm.Model
	GithubInstallationId int64
	GithubAppId          int64
	AccountId            int
	Login                string
	Repo                 string
	State                GithubAppInstallState
}

type GithubAppInstallationLinkStatus int8

const (
	GithubAppInstallationLinkActive   GithubAppInstallationLinkStatus = 1
	GithubAppInstallationLinkInactive GithubAppInstallationLinkStatus = 2
)

// GithubAppInstallationLink links GitHub App installation Id to Digger's organisation Id
type GithubAppInstallationLink struct {
	gorm.Model
	GithubInstallationId int64 `gorm:"index:idx_github_installation_org,unique"`
	OrganisationId       uint  `gorm:"index:idx_github_installation_org,unique"`
	Organisation         *Organisation
	Status               GithubAppInstallationLinkStatus
}

// GithubDiggerJobLink links GitHub Workflow Job id to Digger's Job Id
type GithubDiggerJobLink struct {
	gorm.Model
	DiggerJobId         string `gorm:"size:50"`
	RepoFullName        string
	GithubJobId         int64
	GithubWorkflowRunId int64
	Status              DiggerJobStatus
}
