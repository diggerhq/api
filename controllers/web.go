package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func MainPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"now": time.Date(2017, 0o7, 0o1, 0, 0, 0, 0, time.UTC),
	})
}
