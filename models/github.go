package models

import "gorm.io/gorm"

type GithubApp struct {
	gorm.Model
	GithubId       int64
	OrganisationId int
	PrivateKey     string
}

type GithubAppInstallState string

const (
	Active  GithubAppInstallState = "active"
	Deleted GithubAppInstallState = "deleted"
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

type GithubAppInstallationLink struct {
	gorm.Model
	GithubInstallationID uint `gorm:"index:idx_github_installation_org,unique"`
	GithubInstallation   *GithubAppInstallation
	OrganisationId       uint `gorm:"index:idx_github_installation_org,unique"`
	Organisation         *Organisation
}
