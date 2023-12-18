package middleware

import (
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

func HttpBasicWebAuth() gin.HandlerFunc {

	return func(c *gin.Context) {
		log.Printf("Restricting access")
		username := os.Getenv("HTTP_BASIC_AUTH_USERNAME")
		password := os.Getenv("HTTP_BASIC_AUTH_PASSWORD")
		if username == "" || password == "" {
			c.Error(fmt.Errorf("configuration error: HTTP Basic Auth configured but username or password not set"))
		}
		gin.BasicAuth(gin.Accounts{
			username: password,
		})(c)
		setDefaultOrganisationId(c)
		c.Next()
	}
}

func setDefaultOrganisationId(c *gin.Context) {
	orgNumberOne, err := models.DB.GetOrganisation(models.DEFAULT_ORG_NAME)
	if err != nil {
		c.Error(fmt.Errorf("Error fetching default organisation please check your configuration"))
	}
	c.Set(ORGANISATION_ID_KEY, orgNumberOne.ID)
}

func HttpBasicApiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.String(http.StatusForbidden, "No Authorization header provided")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.String(http.StatusForbidden, "Could not find bearer token in Authorization header")
			c.Abort()
			return
		}

		if token == os.Getenv("BEARER_AUTH_TOKEN") {
			setDefaultOrganisationId(c)
			c.Next()
		}
		return
	}
}
