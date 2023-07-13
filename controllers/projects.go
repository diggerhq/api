package controllers

import (
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func FindProjectsForNamespace(c *gin.Context) {
	namespace := c.Param("namespace")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var projects []models.Project

	err := models.DB.Preload("Organisation").Preload("Namespace").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("namespaces.name = ? AND projects.organisation_id = ?", namespace, orgId).Find(&projects).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return
	}

	response := make([]interface{}, 0)

	for _, p := range projects {
		marshalled := p.MapToJsonStruct()
		response = append(response, marshalled)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while marshalling response")
		return
	}

	c.JSON(http.StatusOK, response)

}

func FindProjectsForOrg(c *gin.Context) {
	requestedOrganisation := c.Param("organisation")
	loggedInOrganisation, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var org models.Organisation
	err := models.DB.Where("name = ?", requestedOrganisation).First(&org).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Could not find organisation: "+requestedOrganisation)
		} else {
			c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		}
		return
	}

	if org.ID != loggedInOrganisation {
		log.Printf("Organisation ID %v does not match logged in organisation ID %v", org.ID, loggedInOrganisation)
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var projects []models.Project

	err = models.DB.Preload("Organisation").Preload("Namespace").
		Joins("LEFT JOIN namespaces ON projects.namespace_id = namespaces.id").
		Joins("LEFT JOIN organisations ON projects.organisation_id = organisations.id").
		Where("projects.organisation_id = ?", org.ID).Find(&projects).Error

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while fetching database")
		return
	}

	response := make([]interface{}, 0)

	for _, p := range projects {
		marshalled := p.MapToJsonStruct()
		response = append(response, marshalled)
	}

	if err != nil {
		c.String(http.StatusInternalServerError, "Unknown error occurred while marshalling response")
		return
	}

	c.JSON(http.StatusOK, response)
}

func ReportProjectsForOrg(c *gin.Context) {
	return // TODO
}
