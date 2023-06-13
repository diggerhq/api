package controllers

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreatePolicyInput struct {
	Policy string
}

func FindPolicy(c *gin.Context) {
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	var policy models.Policy
	models.DB.Take(&policy, "namespace=? AND project_name=? AND organisation_id= ?", namespace, projectName, 1)
	c.JSON(http.StatusOK, policy.Policy)
}

// TODO: Check for policy validation endpoint
func UpdatePolicy(c *gin.Context) {
	// Validate input
	var input CreatePolicyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	namespace := c.Param("namespace")
	projectName := c.Param("projectName")

	policy := models.Policy{}
	result := models.DB.Take(&policy, models.Policy{
		Namespace:      namespace,
		ProjectName:    projectName,
		Type:           "access",
		CreatedBy:      1,
		OrganisationId: 1,
	})
	if result.RowsAffected == 0 {
		models.DB.Create(&models.Policy{
			Namespace:      namespace,
			ProjectName:    projectName,
			Type:           "access",
			CreatedBy:      1,
			OrganisationId: 1,
			Policy:         input.Policy,
		})
	} else {
		result.Update("policy", input.Policy)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
