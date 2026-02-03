package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yeegeek/go-rest-api-starter/internal/contextutil"
	"github.com/yeegeek/go-rest-api-starter/internal/errors"
)

// RequireRole returns a middleware that checks if the user has the specified role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !contextutil.HasRole(c, role) {
			c.JSON(http.StatusForbidden, errors.Forbidden("insufficient permissions"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin returns a middleware that checks if the user is an admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}
