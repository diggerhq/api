package models

import "gorm.io/gorm"

type GithubApp struct {
	gorm.Model
	GithubId     int64
	Name         string
	GithubAppUrl string
}

type GithubAppInstallStatus int

const (
	GithubAppInstallActive  GithubAppInstallStatus = 1
	GithubAppInstallDeleted GithubAppInstallStatus = 2
)

type GithubAppInstallation struct {
	gorm.Model
	GithubInstallationId int64
	GithubAppId          int64
	AccountId            int
	Login                string
	Repo                 string
	Status               GithubAppInstallStatus
}

const (
	GithubAppInstallationLinkActive   string = "active"
	GithubAppInstallationLinkInactive string = "inactive"
)

// GithubAppInstallationLink links GitHub App installation Id to Digger's organisation Id
type GithubAppInstallationLink struct {
	gorm.Model
	GithubInstallationId int64 `gorm:"index:idx_github_installation_org,unique"`
	OrganisationId       uint  `gorm:"index:idx_github_installation_org,unique"`
	Organisation         *Organisation
	Status               string
}
