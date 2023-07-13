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

type Project struct {
	gorm.Model
	Name           string `gorm:"uniqueIndex:idx_project"`
	OrganisationID uint   `gorm:"uniqueIndex:idx_project"`
	Organisation   Organisation
	NamespaceID    uint `gorm:"uniqueIndex:idx_project"`
	Namespace      Namespace
}

func (p *Project) MapToJsonStruct() interface{} {
	return struct {
		Id             uint   `json:"id"`
		Name           string `json:"name"`
		OrganisationID uint   `json:"organisationId"`
		Organisation   string `json:"organisation"`
		NamespaceID    uint   `json:"namespaceId"`
		Namespace      string `json:"namespace"`
	}{
		Id:             p.ID,
		Name:           p.Name,
		OrganisationID: p.OrganisationID,
		NamespaceID:    p.NamespaceID,
		Organisation:   p.Organisation.Name,
		Namespace:      p.Namespace.Name,
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
