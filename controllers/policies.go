package controllers

import (
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
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
	query := JoinedOrganisationNamespaceProjectQuery()

	if namespace != "" && projectName != "" {
		err := query.
			Where("namespaces.name = ? AND projects.name = ?", namespace, projectName).
			First(&policy).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, fmt.Sprintf("Could not find policy for namespace %v and project name %v", namespace, projectName))
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else if organisation != "" {
		err := query.
			Where("organisations.name = ? AND (namespaces.id IS NULL AND projects.id IS NULL)", organisation).
			First(&policy).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.String(http.StatusNotFound, "Could not find policy for organisation: "+organisation)
			} else {
				c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
			}
			return
		}
	} else {
		c.String(http.StatusBadRequest, "Should pass either organisation or namespace + project name")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, policy.Policy)
}

func JoinedOrganisationNamespaceProjectQuery() *gorm.DB {
	return models.DB.Preload("Organisation").Preload("Namespace").Preload("Project").
		Joins("LEFT JOIN namespaces ON policies.namespace_id = namespaces.id").
		Joins("LEFT JOIN projects ON policies.project_id = projects.id")
}

func UpsertPolicyForOrg(c *gin.Context) {
	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	organisation := c.Param("organisation")

	org := models.Organisation{}
	orgResult := models.DB.Where("name = ?", organisation).Take(&org)
	if orgResult.RowsAffected == 0 {
		c.String(http.StatusNotFound, "Could not find organisation: "+organisation)
		return
	}

	policy := models.Policy{}

	policyResult := models.DB.Where("organisation_id = ? AND (namespace_id IS NULL AND project_id IS NULL)", org.ID).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: org.ID,
			Type:           "access",
			Policy:         string(policyData),
		}).Error

		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func UpsertPolicyForNamespaceAndProject(c *gin.Context) {

	orgID, exists := c.Get("organisationID")

	if !exists {
		c.String(http.StatusUnauthorized, "Not authorized")
		return
	}

	orgID = orgID.(uint)

	// Validate input
	policyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		// Handle the error
		c.String(http.StatusInternalServerError, "Error reading request body")
		return
	}
	namespace := c.Param("namespace")
	projectName := c.Param("projectName")
	namespaceModel := models.Namespace{}
	namespaceResult := models.DB.Where("name = ?", namespace).Take(&namespaceModel)
	if namespaceResult.RowsAffected == 0 {
		c.String(http.StatusNotFound, "Could not find namespace: "+namespace)
		return
	}

	projectModel := models.Project{}
	projectResult := models.DB.Where("name = ?", projectName).Take(&projectModel)
	if projectResult.RowsAffected == 0 {
		c.String(http.StatusNotFound, "Could not find project: "+projectName)
		return
	}

	var policy models.Policy

	policyResult := models.DB.Where("organisation_id = ? AND namespace_id = ? AND project_id = ?", namespaceModel.ID, projectModel.ID).Take(&policy)

	if policyResult.RowsAffected == 0 {
		err := models.DB.Create(&models.Policy{
			OrganisationID: orgID.(uint),
			NamespaceID:    &namespaceModel.ID,
			ProjectID:      &projectModel.ID,
			Type:           "access",
			Policy:         string(policyData),
		}).Error
		if err != nil {
			log.Printf("Error creating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error creating policy")
			return
		}
	} else {
		err := policyResult.Update("policy", string(policyData)).Error
		if err != nil {
			log.Printf("Error updating policy: %v", err)
			c.String(http.StatusInternalServerError, "Error updating policy")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
