package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/middleware"
	"digger.dev/cloud/models"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type WebController struct {
	Config *config.Config
}

func (web *WebController) MainPage(c *gin.Context) {
	url := web.Config.Get("FRONTEGG_URL")
	clientId := web.Config.Get("FRONTEGG_CLIENT_ID")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"FronteggUrl":      url,
		"FronteggClientId": clientId,
	})
}

func (web *WebController) ProjectsPage(c *gin.Context) {
	requestedOrganisation := c.Param("organisation")
	loggedInOrganisation, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	fmt.Printf("read org id %v\n", loggedInOrganisation)

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

	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}
