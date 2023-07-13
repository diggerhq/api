package models

import (
	"gorm.io/gorm"
)

type GitHubAppState string

const (
	Active  GitHubAppState = "ACTIVE"
	Deleted GitHubAppState = "DELETED"
)

type GitHubAppInstallation struct {
	gorm.Model
	InstallationId int
	AccountId      int
	Login          string
	Repo           string
	State          GitHubAppState
}
