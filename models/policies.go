package models

import "gorm.io/gorm"

type Policy struct {
	gorm.Model
	Project        *Project
	ProjectID      *uint
	Policy         string
	Type           string
	CreatedBy      *User
	CreatedByID    *uint
	Organisation   Organisation
	OrganisationID uint
	Namespace      *Namespace
	NamespaceID    *uint
}
