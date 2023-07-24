package controllers

import (
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
