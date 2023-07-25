package controllers

import (
	"digger.dev/cloud/config"
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
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
	p1 := models.Project{Name: "project1"}
	p2 := models.Project{Name: "project2"}
	projects := make([]models.Project, 0)
	projects = append(projects, p1)
	projects = append(projects, p2)
	c.HTML(http.StatusOK, "projects.tmpl", gin.H{
		"Projects": projects,
	})
}
