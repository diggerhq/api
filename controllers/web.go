package controllers

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)

func MainPage(c *gin.Context) {
	url := os.Getenv("FRONTEGG_URL")
	clientId := os.Getenv("FRONTEGG_CLIENT_ID")
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"FronteggUrl":      url,
		"FronteggClientId": clientId,
	})
}

func ProjectsPage(c *gin.Context) {

	p1 := models.Project{Name: "project1"}
	p2 := models.Project{Name: "project2"}
	projects := make([]models.Project, 0)
	projects = append(projects, p1)
	projects = append(projects, p2)
	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}
