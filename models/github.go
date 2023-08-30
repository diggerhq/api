package models

import "gorm.io/gorm"

type GithubApp struct {
	gorm.Model
	GithubId       int64  `json:"githubId"`
	OrganisationId int    `json:"organisationId"`
	PrivateKey     string `json:"privateKey"`
}

type GithubAppInstallState string

const (
	Active  GithubAppInstallState = "active"
	Deleted GithubAppInstallState = "deleted"
)

type GithubAppInstallation struct {
	gorm.Model
	GithubInstallationId int64 `json:"githubInstallationId"`
	GithubAppId          int64 `json:"githubAppId"`
	AccountId            int
	Login                string
	Repo                 string
	State                GithubAppInstallState `gorm:"type:enum('active', 'deleted')"`
}
