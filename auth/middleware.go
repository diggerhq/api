package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/bmdavis419/golang-auth0-example/config"
)

type AuthMiddleware struct {
	config config.EnvVars
}

func NewAuthMiddleware(config config.EnvVars) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

		issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
		if err != nil {
			log.Fatalf("Failed to parse the issuer url: %v", err)
		}

		provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

		jwtValidator, err := validator.New(
			provider.KeyFunc,
			validator.RS256,
			issuerURL.String(),
			[]string{os.Getenv("AUTH0_AUDIENCE")},
		)
		if err != nil {
			log.Fatalf("Failed to set up the jwt validator")
		}

		// get the token from the request header
		authHeader, _ := c.Get("Authorization")
		authHeaderParts := strings.Split(fmt.Sprintf("%v", authHeader), " ")
		if len(authHeaderParts) != 2 {
			terminateWithError(http.StatusUnauthorized, "Invalid Authorization Header", c)
		}

		// Validate the token
		tokenInfo, err := jwtValidator.ValidateToken(c.Request.Context(), authHeaderParts[1])
		if err != nil {
			fmt.Println(err)
			terminateWithError(http.StatusUnauthorized, "token is not valid", c)
		}

		fmt.Println(tokenInfo)

		// Go to next middleware:
		c.Next()
	}
}

func terminateWithError(statusCode int, message string, c *gin.Context) {
	c.JSON(statusCode, gin.H{"error": message})
	c.Abort()
}
