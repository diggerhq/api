package models

import "gorm.io/gorm"

type GitHubAppInstallation struct {
	gorm.Model
	InstallationId int
	AccountId      int
	Login          string
	Repo           string
}
