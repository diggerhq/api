package models

import "gorm.io/gorm"

type GithubApp struct {
	gorm.Model
	GithubId       int64  `json:"githubId"`
	OrganisationId int    `json:"organisationId"`
	PrivateKey     string `json:"privateKey"`
}

type GithubAppInstallation struct {
	gorm.Model
	GithubInstallationId int64 `json:"githubInstallationId"`
	GithubAppId          int64 `json:"githubAppId"`
}
