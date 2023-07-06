package middleware

import (
	"digger.dev/cloud/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"os"
	"strings"
)

func SecretCodeAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.Request.Header.Get("x-webhook-secret")
		if secret == "" {
			log.Printf("No x-webhook-secret header provided")
			c.String(http.StatusForbidden, "No x-webhook-secret header provided")
			c.Abort()
			return
		}
		_, err := jwt.Parse(secret, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("WEBHOOK_SECRET")), nil
		})

		if err != nil {
			log.Printf("Error parsing secret: %v", err.Error())
			c.String(http.StatusForbidden, "Invalid x-webhook-secret header provided")
			c.Abort()
			return
		}
		c.Next()
	}
}

func BearerTokenAuth() gin.HandlerFunc {
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

		if strings.HasPrefix(token, "t:") {
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
		} else {
			publicKeyData, err := os.ReadFile("pk.pem")

			if err != nil {
				log.Printf("Error while reading public key: %v", err.Error())
				c.String(http.StatusInternalServerError, "Error occurred while reading public key")
				c.Abort()
				return
			}

			publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
			if err != nil {
				log.Printf("Error while parsing public key: %v", err.Error())
				c.String(http.StatusInternalServerError, "Error occurred while parsing public key")
				c.Abort()
				return
			}

			token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return publicKey, nil
			})

			if err != nil {
				log.Printf("Error while parsing token: %v", err.Error())
				c.String(http.StatusForbidden, "Authorization header is invalid")
				c.Abort()
				return
			}

			if !token.Valid {
				log.Printf("Token is invalid")
				c.String(http.StatusForbidden, "Authorization header is invalid")
				c.Abort()
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if claims.Valid() != nil {
					log.Printf("Token is invalid")
					c.String(http.StatusForbidden, "Authorization header is invalid")
					c.Abort()
					return
				}
				var org models.Organisation
				issuer := claims["iss"]
				if issuer == nil {
					log.Printf("Token is invalid")
					c.String(http.StatusForbidden, "Authorization header is invalid")
					c.Abort()
					return
				}
				issuer = issuer.(string)
				tenantId := claims["tenantId"]
				if tenantId == nil {
					log.Printf("Token is invalid")
					c.String(http.StatusForbidden, "Authorization header is invalid")
					c.Abort()
					return
				}
				tenantId = tenantId.(string)
				err := models.DB.Take(org, "external_source = ? AND external_id = ?", issuer, tenantId).Error
				if err != nil {
					log.Printf("Error while fetching organisation: %v", err.Error())
					c.String(http.StatusForbidden, "Authorization header is invalid")
					c.Abort()
					return
				}
				c.Set(ORGANISATION_ID_KEY, org.ID)

				permissions := claims["permissions"]
				if permissions == nil {
					log.Printf("Token is invalid")
					c.String(http.StatusForbidden, "Authorization header is invalid")
					c.Abort()
					return
				}
				permissions = permissions.([]interface{})
				for _, permission := range permissions.([]interface{}) {
					permission = permission.(string)
					if permission == "digger.all.*" {
						c.Set(ACCESS_LEVEL_KEY, models.AdminPolicyType)
						c.Next()
						return
					}
					if permission == "digger.all.read.*" {
						c.Set(ACCESS_LEVEL_KEY, models.AccessPolicyType)
						c.Next()
						return
					}
				}
			} else {
				log.Printf("Token is invalid")
				c.String(http.StatusForbidden, "Authorization header is invalid")
				c.Abort()
				return
			}
		}

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
