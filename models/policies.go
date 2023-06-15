package models

type Policy struct {
	ID             uint   `json:"id" gorm:"primary_key"`
	Organisation   string `json:"organisation"`
	Namespace      string `json:"namespace"`
	ProjectName    string `json:"project_name"` // The digger project name
	Policy         string `json:"policy"`
	Type           string `json:"type"`            // TODO: Make this an Enum(access,login, ...)
	CreatedBy      int    `json:"created_by"`      // TODO: Make this an fk to user
	OrganisationId int    `json:"organisation_id"` // TODO: Make this an fk to organisation
}
