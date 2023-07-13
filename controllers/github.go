package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func GitHubAppCallback() func(c *gin.Context) {
	return func(c *gin.Context) {
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading request body")
			return
		}
		c.Header("Content-Type", "application/json")
		fmt.Print("---------- github app callback ---------------- ")
		fmt.Printf(string(requestBody))
		c.JSON(200, "ok")
	}
}

func GitHubAppWebHook() func(c *gin.Context) {
	return func(c *gin.Context) {
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error reading request body")
			return
		}
		c.Header("Content-Type", "application/json")
		fmt.Print("---------- github app webhook ---------------- ")
		fmt.Printf(string(requestBody))
		c.JSON(200, "ok")
	}
}
