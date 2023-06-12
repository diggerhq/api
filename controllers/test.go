package controllers

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

type CreateTestInput struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

func FindTest(c *gin.Context) {
	var books []models.Test
	models.DB.Find(&books)

	c.JSON(http.StatusOK, gin.H{"data": books})
}
func CreateTest(c *gin.Context) {
	// Validate input
	var input CreateTestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create book
	book := models.Test{Title: input.Title, Author: input.Author}
	models.DB.Create(&book)

	c.JSON(http.StatusOK, gin.H{"data": book})
}
