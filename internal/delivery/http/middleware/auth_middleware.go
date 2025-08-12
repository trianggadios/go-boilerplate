package middleware

import (
	"boilerplate-go/pkg/jwt"
	"boilerplate-go/pkg/response"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required", "missing authorization header")
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format", "expected Bearer token")
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := jwt.ValidateToken(token, secretKey)
		if err != nil {
			response.Unauthorized(c, "Invalid token", err.Error())
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
