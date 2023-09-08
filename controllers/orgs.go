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
	org := models.Organisation{
		Name:           json.Name,
		ExternalSource: source,
		ExternalId:     json.TenantId,
	}
	err := models.DB.GormDB.Create(&org).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
