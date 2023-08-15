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
		jsonStruct := p.MapToJsonStruct()
		response = append(response, jsonStruct)
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

type CreateProjectRequest struct {
	Name              string `json:"name"`
	ConfigurationYaml string `json:"configurationYaml"`
}

func ReportProjectsForNamespace(c *gin.Context) {
	var request CreateProjectRequest
	err := c.BindJSON(&request)
	if err != nil {
		log.Printf("Error binding JSON: %v", err)
		return
	}

	namespaceName := c.Param("namespace")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var org models.Organisation

	err = models.DB.Where("id = ?", orgId).First(&org).Error

	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var namespace models.Namespace

	err = models.DB.Where("name = ? AND organisation_id = ?", namespaceName, orgId).First(&namespace).Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			namespace := models.Namespace{
				Name:           namespaceName,
				OrganisationID: org.ID,
				Organisation:   &org,
			}

			err = models.DB.Create(&namespace).Error

			if err != nil {
				log.Printf("Error creating namespace: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating namespace"})
				return
			}
		} else {
			log.Printf("Error fetching namespace: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching namespace"})
			return
		}
	}

	project := models.Project{
		Name:              request.Name,
		ConfigurationYaml: request.ConfigurationYaml,
		NamespaceID:       namespace.ID,
		OrganisationID:    org.ID,
		Namespace:         &namespace,
		Organisation:      &org,
	}

	err = models.DB.Create(&project).Error

	if err != nil {
		log.Printf("Error creating project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating project"})
		return
	}

	c.JSON(http.StatusOK, project.MapToJsonStruct())
}

func RunHistoryForProject(c *gin.Context) {
	namespaceName := c.Param("namespace")
	projectName := c.Param("project")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var org models.Organisation

	err := models.DB.Where("id = ?", orgId).First(&org).Error

	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var namespace models.Namespace

	err = models.DB.Where("name = ? AND organisation_id = ?", namespaceName, orgId).First(&namespace).Error

	if err != nil {
		log.Printf("Error fetching namespace: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching namespace"})
		return
	}

	var project models.Project

	err = models.DB.Where("name = ? AND namespace_id = ? AND organisation_id", projectName, namespace.ID, org.ID).First(&project).Error

	if err != nil {
		log.Printf("Error fetching project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project"})
		return
	}

	var runHistory []models.ProjectRun

	err = models.DB.Where("project_id = ?", project.ID).Find(&runHistory).Error

	if err != nil {
		log.Printf("Error fetching run history: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching run history"})
		return
	}

	response := make([]interface{}, 0)

	for _, r := range runHistory {
		response = append(response, r.MapToJsonStruct())
	}

	c.JSON(http.StatusOK, response)
}

type CreateProjectRunRequest struct {
	StartedAt int64  `json:"startedAt"`
	EndedAt   int64  `json:"endedAt"`
	Status    string `json:"status"`
	Command   string `json:"command"`
	Output    string `json:"output"`
}

func CreateRunForProject(c *gin.Context) {
	namespaceName := c.Param("namespace")
	projectName := c.Param("project")
	orgId, exists := c.Get(middleware.ORGANISATION_ID_KEY)

	if !exists {
		c.String(http.StatusForbidden, "Not allowed to access this resource")
		return
	}

	var org models.Organisation

	err := models.DB.Where("id = ?", orgId).First(&org).Error

	if err != nil {
		log.Printf("Error fetching organisation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching organisation"})
		return
	}

	var namespace models.Namespace

	err = models.DB.Where("name = ? AND organisation_id = ?", namespaceName, orgId).First(&namespace).Error

	if err != nil {
		log.Printf("Error fetching namespace: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching namespace"})
		return
	}

	var project models.Project

	err = models.DB.Where("name = ? AND namespace_id = ? AND organisation_id", projectName, namespace.ID, org.ID).First(&project).Error

	if err != nil {
		log.Printf("Error fetching project: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching project"})
		return
	}

	var request CreateProjectRunRequest

	err = c.BindJSON(&request)

	if err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error binding JSON"})
		return
	}

	run := models.ProjectRun{
		StartedAt: request.StartedAt,
		EndedAt:   request.EndedAt,
		Status:    request.Status,
		Command:   request.Command,
		Output:    request.Output,
		ProjectID: project.ID,
		Project:   &project,
	}

	err = models.DB.Create(&run).Error

	if err != nil {
		log.Printf("Error creating run: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating run"})
		return
	}

	c.JSON(http.StatusOK, run.MapToJsonStruct())
}
