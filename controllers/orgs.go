package controllers

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TenantCreatedEvent struct {
	TenantId string `json:"tenantId,omitempty"`
	Name     string `json:"name,omitempty"`
}

func CreateFronteggOrgFromWebhook(c *gin.Context) {
	var json TenantCreatedEvent

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	source := c.GetHeader("x-tenant-source")

	_, err := models.DB.CreateOrganisation(json.Name, source, json.TenantId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organisation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
