package middleware

import (
	"net/http"
	"strings"

	"FreeTranslate/internal/platform/config"
	"FreeTranslate/internal/platform/gwe"

	"github.com/gin-gonic/gin"
)

// AuthToken 验证 Authorization Bearer Token
// 请求头格式：Authorization: Bearer <token>
func AuthToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			gwe.ErrorJSON(c, http.StatusUnauthorized, 40101, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			gwe.ErrorJSON(c, http.StatusUnauthorized, 40102, "invalid authorization header format")
			c.Abort()
			return
		}

		token := parts[1]
		if token != config.Config.APIToken {
			gwe.ErrorJSON(c, http.StatusUnauthorized, 40103, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}