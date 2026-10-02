package rpc

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const BearerSchema = "Bearer"

func validateToken(configBearerToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := tokenFromAuthHeader(c.Request)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"msg": err.Error()})
			c.Abort()
			return
		}

		tokenDecoded, err := base64.StdEncoding.DecodeString(tokenString)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"msg": fmt.Sprintf("decoding token failed : %v", err)})
			c.Abort()
			return
		}

		if string(tokenDecoded) != configBearerToken {
			c.JSON(http.StatusForbidden, gin.H{"msg": "invalid token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func tokenFromAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("'Authorization' header not found")
	}

	authHeaderParts := strings.Fields(authHeader)
	if len(authHeaderParts) != 2 || authHeaderParts[0] != BearerSchema {
		return "", fmt.Errorf("'Authorization' header format must be Bearer {token}")
	}

	return authHeaderParts[1], nil
}
