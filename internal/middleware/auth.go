package middleware

import (
	"net/http"
	"strings"

	"proem-backend/pkg/token"

	"github.com/gin-gonic/gin"
)

const CtxUserIDKey = "userID"

func AuthMiddleware(jwtMgr *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		claims, err := jwtMgr.Verify(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Next()
	}
}
