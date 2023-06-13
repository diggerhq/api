package controllers

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreatePolicyInput struct {
	Namespace      string
	ProjectName    string
	Policy         string
	Type           string
	CreatedBy      int
	OrganisationId int
}

func FindPolicy(c *gin.Context) {
	var policies []models.Policy
	models.DB.Find(&policies, "organisation_id= ?", 1)
	c.JSON(http.StatusOK, gin.H{"data": policies})
}

func CreatePolicy(c *gin.Context) {
	// Validate input
	var input CreatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create policy
	policy := models.Policy{
		Namespace:      input.Namespace,
		ProjectName:    input.ProjectName,
		Policy:         input.Policy,
		Type:           "access",
		CreatedBy:      1,
		OrganisationId: 1,
	}
	models.DB.Create(&policy)

	c.JSON(http.StatusOK, gin.H{"data": policy})
}
