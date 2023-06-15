package controllers

import (
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
)

type CreatePolicyInput struct {
	Policy string
}

func FindPolicy(c *gin.Context) {
	namespace := c.Param("namespace")
	organisation := c.Param("organisation")
	projectName := c.Param("projectName")
	var policy models.Policy

	if namespace != "" {
		if err := models.DB.Take(&policy, "namespace=? AND project_name=? AND organisation_id= ?", namespace, projectName, 1).Error; err != nil {
			fmt.Printf("Error during namespace query %v", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, "Could not find policy for namespace: "+namespace)
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else if organisation != "" {
		if err := models.DB.Take(&policy, "organisation=? AND project_name=? AND organisation_id= ?", organisation, projectName, 1).Error; err != nil {
			fmt.Printf("Error during namespace query %v", err)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, "Could not find policy for organisation: "+organisation)
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else {
		c.String(http.StatusBadRequest, "Should pass either organisation or namespace")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

// TODO: Check for policy validation endpoint
func UpdatePolicy(c *gin.Context) {
	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	organisation := c.Param("organisation")

	policy := models.Policy{}
	result := models.DB.Take(&policy, models.Policy{
		Namespace:      namespace,
		Organisation:   organisation,
		ProjectName:    projectName,
		Type:           "access",
		CreatedBy:      1,
		OrganisationId: 1,
	})
	if result.RowsAffected == 0 {
		models.DB.Create(&models.Policy{
			Namespace:      namespace,
			Organisation:   organisation,
			ProjectName:    projectName,
			Type:           "access",
			CreatedBy:      1,
			OrganisationId: 1,
			Policy:         string(policyData),
		})
	} else {
		result.Update("policy", string(policyData))
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
