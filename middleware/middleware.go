package middleware

import (
	"digger.dev/cloud/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

func BasicBearerTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		if auth == "" {
			c.String(http.StatusForbidden, "No Authorization header provided")
			c.Abort()
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth {
			c.String(http.StatusForbidden, "Could not find bearer token in Authorization header")
			c.Abort()
			return
		}
		var dbToken models.Token

		tokenResults := models.DB.Take(&dbToken, "value = ?", token)

		if tokenResults.RowsAffected == 0 {
			c.String(http.StatusForbidden, "Invalid bearer token")
			c.Abort()
			return
		}

		if tokenResults.Error != nil {
			log.Printf("Error while fetching token from database: %v", tokenResults.Error)
			c.String(http.StatusInternalServerError, "Error occurred while fetching database")
			c.Abort()
			return
		}
		c.Set(ORGANISATION_ID_KEY, dbToken.OrganisationID)
		c.Set(ACCESS_LEVEL_KEY, dbToken.Type)

		c.Next()
	}
}

func AccessLevel(allowedAccessLevels ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessLevel := c.GetString(ACCESS_LEVEL_KEY)
		for _, allowedAccessLevel := range allowedAccessLevels {
			if accessLevel == allowedAccessLevel {
				c.Next()
				return
			}
		}
		c.String(http.StatusForbidden, "Not allowed to access this resource with this access level")
		c.Abort()
	}
}

const ORGANISATION_ID_KEY = "organisation_ID"
const ACCESS_LEVEL_KEY = "access_level"
