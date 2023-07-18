package models

import (
	"gorm.io/gorm"
)

type Organisation struct {
	gorm.Model
	Name           string `gorm:"uniqueIndex:idx_organisation"`
	ExternalSource string `gorm:"uniqueIndex:idx_external_source"`
	ExternalId     string `gorm:"uniqueIndex:idx_external_source"`
}

type Namespace struct {
	gorm.Model
	Name           string `gorm:"uniqueIndex:idx_org_namespace"`
	OrganisationID uint   `gorm:"uniqueIndex:idx_org_namespace"`
	Organisation   Organisation
}

type ProjectRun struct {
	gorm.Model
	ProjectID       uint
	Project         Project
	StartedAt       int64
	EndedAt         int64
	Status          string
	Command         string
	TerraformOutput string
}

func (p *ProjectRun) MapToJsonStruct() interface{} {
	return struct {
		Id              uint   `json:"id"`
		ProjectID       uint   `json:"projectId"`
		ProjectName     string `json:"projectName"`
		StartedAt       int64  `json:"startedAt"`
		EndedAt         int64  `json:"endedAt"`
		Status          string `json:"status"`
		Command         string `json:"command"`
		TerraformOutput string `json:"terraformOutput"`
	}{
		Id:              p.ID,
		ProjectID:       p.ProjectID,
		ProjectName:     p.Project.Name,
		StartedAt:       p.StartedAt,
		EndedAt:         p.EndedAt,
		Status:          p.Status,
		Command:         p.Command,
		TerraformOutput: p.TerraformOutput,
	}
}

type Project struct {
	gorm.Model
	Name              string `gorm:"uniqueIndex:idx_project"`
	OrganisationID    uint   `gorm:"uniqueIndex:idx_project"`
	Organisation      Organisation
	NamespaceID       uint `gorm:"uniqueIndex:idx_project"`
	Namespace         Namespace
	ConfigurationYaml string
}

func (p *Project) MapToJsonStruct() interface{} {
	return struct {
		Id               uint   `json:"id"`
		Name             string `json:"name"`
		OrganisationID   uint   `json:"organisationId"`
		OrganisationName string `json:"organisationName"`
		NamespaceID      uint   `json:"namespaceId"`
		NamespaceName    string `json:"namespaceName"`
	}{
		Id:               p.ID,
		Name:             p.Name,
		OrganisationID:   p.OrganisationID,
		NamespaceID:      p.NamespaceID,
		OrganisationName: p.Organisation.Name,
		NamespaceName:    p.Namespace.Name,
	}
}

type Token struct {
	gorm.Model
	Value          string `gorm:"uniqueIndex:idx_token"`
	OrganisationID uint
	Organisation   Organisation
	Type           string
}

const (
	AccessPolicyType = "access"
	AdminPolicyType  = "admin"
)
