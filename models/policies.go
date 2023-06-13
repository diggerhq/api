package models

type Policy struct {
	ID             uint   `json:"id" gorm:"primary_key"`
	Namespace      string `json:"namespace"`
	ProjectName    string `json:"project_name"`
	Policy         string `json:"policy"`
	Type           string `json:"type"`            // TODO: Make this an Enum(access,login, ...)
	CreatedBy      int    `json:"created_by"`      // TODO: Make this an fk to user
	OrganisationId int    `json:"organisation_id"` // TODO: Make this an fk to organisation
}
